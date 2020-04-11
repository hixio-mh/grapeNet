// 连接对象
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/10

package grapeNet

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"

	utils "github.com/koangel/grapeNet/Utils"
)

type TcpConn struct {
	cm.Conn
	ownerNet *TCPNetwork
	TConn    net.Conn    // tcp连接
	UserData interface{} // 用户对象
	LastPing time.Time

	send    chan []byte
	process chan *cm.QPackItem // 单独的一个数据包

	CryptKey []byte

	IsClosed int32

	removeOnce sync.Once
	closeSock  sync.Once

	remoteAddr string

	ReadTime  time.Duration
	WriteTime time.Duration
}

const (
	ReadWaitPing = 65 * time.Second
	WriteTicker  = 45 * time.Second

	queueCount    = 2048
	maxPacketSize = 6 * 1024 * 1024 // 6兆数据包最大
)

const (
	RMStream     = iota // 使用流算法读写数据包
	RMReadFull          // 使用固定算法读写数据包 固定算法 包头4字节为固定长度，其余位置为数据包payload
	RMStringLine        // 通过使用\n来进行文本拆分
)

//////////////////////////////////////
// 创建新连接

func EmptyConn(ctype int) *TcpConn {
	newConn := &TcpConn{
		LastPing: time.Now(),
		IsClosed: 0,

		CryptKey: []byte{},

		send:    make(chan []byte, queueCount),
		process: make(chan *cm.QPackItem, queueCount),

		WriteTime: WriteTicker,
		ReadTime:  ReadWaitPing,
	}

	newConn.Ctx, newConn.Cancel = context.WithCancel(context.Background())
	newConn.Once = new(sync.Once)
	newConn.Wg = new(sync.WaitGroup)
	newConn.SessionId = cm.CreateUUID(ctype)
	newConn.Type = ctype

	return newConn
}

func NewConn(tn *TCPNetwork, conn net.Conn, UData interface{}) *TcpConn {
	newConn := EmptyConn(cm.ESERVER_TYPE)

	newConn.TConn = conn
	newConn.ownerNet = tn
	newConn.UserData = UData
	newConn.remoteAddr = conn.RemoteAddr().String()

	return newConn
}

func NewDial(tn *TCPNetwork, addr string, UData interface{}) (conn *TcpConn, err error) {
	err = nil
	conn = nil
	dconn, derr := net.DialTimeout("tcp", addr, time.Second*30)
	if derr != nil {
		logger.ERRORV(derr)
		err = derr
		return
	}

	tcpConn, ok := dconn.(*net.TCPConn)
	if ok {
		tcpConn.SetKeepAlive(true)
	}

	conn = EmptyConn(cm.ECLIENT_TYPE)
	conn.ownerNet = tn
	conn.TConn = dconn
	conn.UserData = UData
	conn.remoteAddr = dconn.RemoteAddr().String()

	return
}

//////////////////////////////////////////////
// 成员函数
func (c *TcpConn) SetReadTime(d time.Duration) {
	c.ReadTime = d
}

func (c *TcpConn) SetWriteTime(w time.Duration) {
	c.WriteTime = w
}

func (c *TcpConn) SetUserData(user interface{}) {
	c.UserData = user
}

func (c *TcpConn) GetUserData() interface{} {
	return c.UserData
}

func (c *TcpConn) GetNetConn() net.Conn {
	return c.TConn
}

func (c *TcpConn) RemoteAddr() string {
	return c.remoteAddr
}

func (c *TcpConn) startProc() {

	go c.writePump()

	switch c.ownerNet.RecvMode {
	case RMReadFull:
		go c.recvPumpFull()
	case RMStringLine:
		go c.recvPumpString()
	default:
		go c.recvPump()
	}

	if HandlerProc > 0 {
		go c.handlerPump()
	}
}

func (c *TcpConn) handlerPump() {
	defer func() {
		if p := recover(); p != nil {
			stacks := utils.PanicTrace(4)
			panic := fmt.Sprintf("recover panics: %v call:%v", p, string(stacks))
			logger.ERROR(panic)

			if c.ownerNet.Panic != nil {
				c.ownerNet.Panic(c, panic)
			}
		}

		c.Wg.Done()
		logger.INFOV("addr:", c.remoteAddr, ",", c.SessionId, " handle full Pump defer done!!!")
	}()

	c.Wg.Add(1)
	for {
		select {
		case <-c.Ctx.Done():
			logger.INFO("%v %v session handler done...", c.remoteAddr, c.SessionId)
			return
		case item := <-c.process:
			{
				if atomic.LoadInt32(&c.IsClosed) == 1 {
					return
				}

				if c.ownerNet.OnHandler != nil {
					c.ownerNet.OnHandler(c, item.Payload)
				}
			}
		}
	}
}

func (c *TcpConn) recvPump() {
	defer func() {
		if p := recover(); p != nil {
			stacks := utils.PanicTrace(4)
			panic := fmt.Sprintf("recover panics: %v call:%v", p, string(stacks))
			logger.ERROR(panic)

			if c.ownerNet.Panic != nil {
				c.ownerNet.Panic(c, panic)
			}
		}

		c.Cancel() // 结束
		c.Wg.Wait()
		c.Close() // 关闭SOCKET
		if c.ownerNet != nil {
			c.ownerNet.RemoveSession(c.SessionId) // 删除
		}
		logger.INFOV("addr:", c.remoteAddr, ",", c.SessionId, " read Pump defer done!!!")
	}()

	var buffer = make([]byte, 65535)
	var lStream stream.BufferIO

	for {
		c.TConn.SetReadDeadline(time.Now().Add(c.ReadTime))
		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		rn, err := c.TConn.Read(buffer)
		if err != nil {
			logger.ERROR("Session %v %v Recv Error:%v", c.remoteAddr, c.SessionId, err)
			return
		}

		if rn == 0 {
			logger.ERROR("Session %v %v Recv Len:%v", c.remoteAddr, c.SessionId, rn)
			return
		}

		if c.ownerNet.OnHandler == nil {
			logger.ERROR("%v Handler is Null,Closed...", c.remoteAddr)
			return
		}

		if lStream.Write(buffer, rn) == -1 {
			logger.ERROR("Session %v %v Recv Error:Packet size is too big...", c.remoteAddr, c.SessionId)
			return
		}

		upak, err := c.ownerNet.Unpackage(c, &lStream) // 调用解压行为
		if err == nil {
			for _, v := range upak {
				if c.ownerNet.SendPong(c, v) {
					continue
				}

				if HandlerProc <= 0 {
					if c.ownerNet.OnHandler != nil {
						c.ownerNet.OnHandler(c, v[4:])
					}
				} else {
					if c.process != nil {
						c.process <- &cm.QPackItem{
							Length:  int32(len(v[4:])),
							Payload: v[4:],
						}
					}
				}
			}
		}
	}
}

func (c *TcpConn) recvPumpString() {
	defer func() {
		if p := recover(); p != nil {
			stacks := utils.PanicTrace(4)
			panic := fmt.Sprintf("recover panics: %v call:%v", p, string(stacks))
			logger.ERROR(panic)

			if c.ownerNet.Panic != nil {
				c.ownerNet.Panic(c, panic)
			}
		}

		c.Cancel() // 结束
		c.Wg.Wait()
		c.Close() // 关闭SOCKET
		if c.ownerNet != nil {
			c.ownerNet.RemoveSession(c.SessionId) // 删除
		}

		logger.INFOV("addr:", c.remoteAddr, ",", c.SessionId, " read string Pump defer done!!!")
	}()

	readerLine := bufio.NewScanner(c.TConn)
	readerLine.Split(bufio.ScanLines)
	for readerLine.Scan() {
		c.TConn.SetReadDeadline(time.Now().Add(c.ReadTime))
		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		if c.ownerNet.OnHandler == nil {
			logger.ERROR("%v Handler is Null,Closed...", c.remoteAddr)
			return
		}

		payload := c.ownerNet.Decrypt(readerLine.Bytes(), c.CryptKey)
		if c.ownerNet.SendPong(c, payload) {
			continue
		}

		if HandlerProc <= 0 {
			if c.ownerNet.OnHandler != nil {
				c.ownerNet.OnHandler(c, payload)
			}
		} else {
			if c.process != nil {
				c.process <- &cm.QPackItem{
					Length:  int32(len(payload)),
					Payload: payload,
				}
			}
		}
	}
}

func (c *TcpConn) recvPumpFull() {
	defer func() {
		if p := recover(); p != nil {
			stacks := utils.PanicTrace(4)
			panic := fmt.Sprintf("recover panics: %v call:%v", p, string(stacks))
			logger.ERROR(panic)

			if c.ownerNet.Panic != nil {
				c.ownerNet.Panic(c, panic)
			}
		}

		c.Cancel() // 结束
		c.Wg.Wait()
		c.Close() // 关闭SOCKET
		if c.ownerNet != nil {
			c.ownerNet.RemoveSession(c.SessionId) // 删除
		}

		logger.INFOV("addr:", c.remoteAddr, ",", c.SessionId, " read full Pump defer done!!!")
	}()

	for {
		c.TConn.SetReadDeadline(time.Now().Add(c.ReadTime))
		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		lenBytes := make([]byte, 4)
		rn, err := io.ReadFull(c.TConn, lenBytes)
		if err != nil {
			logger.ERROR("Session %v %v Recv Error:%v", c.remoteAddr, c.SessionId, err)
			return
		}

		if rn == 0 {
			logger.ERROR("Session %v %v Recv Len:%v", c.remoteAddr, c.SessionId, rn)
			return
		}

		headerLen := binary.LittleEndian.Uint32(lenBytes)
		payload := make([]byte, headerLen)
		rn, err = io.ReadFull(c.TConn, payload)
		if err != nil {
			logger.ERROR("Session %v %v Recv Error:%v", c.remoteAddr, c.SessionId, err)
			return
		}

		if rn == 0 {
			logger.ERROR("Session %v %v Recv Len:%v", c.remoteAddr, c.SessionId, rn)
			return
		}

		if c.ownerNet.OnHandler == nil {
			logger.ERROR("%v Handler is Null,Closed...", c.remoteAddr)
			return
		}

		payload = c.ownerNet.Decrypt(payload, c.CryptKey)
		if c.ownerNet.SendPong(c, payload) {
			continue
		}

		if HandlerProc <= 0 {
			if c.ownerNet.OnHandler != nil {
				c.ownerNet.OnHandler(c, payload)
			}
		} else {
			if c.process != nil {
				c.process <- &cm.QPackItem{
					Length:  int32(headerLen),
					Payload: payload,
				}
			}
		}
	}
}

func (c *TcpConn) writePump() {
	heartbeat := time.NewTicker(30 * time.Second)
	c.Wg.Add(1)
	defer func() {
		if p := recover(); p != nil {
			stacks := utils.PanicTrace(4)
			panic := fmt.Sprintf("writePump panics: %v call:%v", p, string(stacks))
			logger.ERROR(panic)

			if c.ownerNet.Panic != nil {
				c.ownerNet.Panic(c, panic)
			}
		}

		heartbeat.Stop()
		c.Wg.Done()
		logger.INFOV("addr:", c.remoteAddr, ",", c.SessionId, " write Pump defer done!!!")
	}()

	for {
		select {
		case <-c.Ctx.Done():
			logger.INFO("%v %v session write done...", c.remoteAddr, c.SessionId)
			return
		case bData, ok := <-c.send:
			if !ok {
				return
			}

			if atomic.LoadInt32(&c.IsClosed) == 1 {
				return
			}

			c.TConn.SetWriteDeadline(time.Now().Add(c.WriteTime))
			if _, err := c.TConn.Write(bData); err != nil {
				logger.ERROR("%v write Pump error:%v !!!", c.remoteAddr, err)
				return
			}
			break
		case <-heartbeat.C:
			if atomic.LoadInt32(&c.IsClosed) == 1 {
				return
			}

			c.TConn.SetWriteDeadline(time.Now().Add(c.WriteTime))
			c.SendDirect([]byte("ping")) // 发送心跳
			break
		}
	}
}

func (c *TcpConn) Send(data []byte) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	var (
		encode []byte
		err    error
	)

	if (c.ownerNet.RecvMode == RMReadFull || c.ownerNet.UseHeaderLen) && c.ownerNet.RecvMode != RMStringLine {
		encode, err = stream.PackerOnce(data, c.ownerNet.Encrypt, c.CryptKey)
		if err != nil {
			return -1
		}
	} else {
		encode = c.ownerNet.Encrypt(data, c.CryptKey)
	}

	select {
	case c.send <- encode:
		return len(data)
	case <-time.After(3 * time.Second):
		break
	default:
		break
	}

	return -1
}

func (c *TcpConn) SendPak(val interface{}) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	if c.ownerNet.Package == nil {
		logger.ERROR("%v Package Func Error,Can't Send...", c.remoteAddr)
		return -1
	}

	pack, err := c.ownerNet.Package(val)
	if err != nil {
		return -1
	}

	return c.Send(pack)
}

func (c *TcpConn) SendDirect(data []byte) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	var (
		encode []byte
		err    error
	)

	if (c.ownerNet.RecvMode == RMReadFull || c.ownerNet.UseHeaderLen) && c.ownerNet.RecvMode != RMStringLine {
		encode, err = stream.PackerOnce(data, c.ownerNet.Encrypt, c.CryptKey)
		if err != nil {
			return -1
		}
	} else {
		encode = c.ownerNet.Encrypt(data, c.CryptKey)
	}

	retry := c.ownerNet.SendRetry
	if retry <= 0 {
		retry = 1
	}

	for i := 0; i < retry; i++ {
		c.TConn.SetWriteDeadline(time.Now().Add(c.WriteTime))
		wn, err := c.TConn.Write(encode)
		if err != nil {
			logger.ERRORV(err)
			continue
		}

		return wn
	}

	return -1
}

func (c *TcpConn) SendPakDirect(val interface{}) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	if c.ownerNet.Package == nil {
		logger.ERROR("%v Package Func Error,Can't Send...", c.remoteAddr)
		return -1
	}

	pack, err := c.ownerNet.Package(val)
	if err != nil {
		return -1
	}

	return c.SendDirect(pack)
}

func (c *TcpConn) CloseSocket() {
	c.closeSock.Do(func() {
		c.Cancel()
		atomic.StoreInt32(&c.IsClosed, 1)
		c.TConn.Close() // 关闭连接
	})
}

func (c *TcpConn) Close() {
	c.Once.Do(func() {
		// 都没连上怎么关
		if c.TConn == nil {
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 0 {
			c.ownerNet.OnClose(c)
			c.CloseSocket()
		}
	})
}

func (c *TcpConn) RemoveData() {
	c.removeOnce.Do(func() {
		if atomic.LoadInt32(&c.IsClosed) == 1 {
			if c.send != nil {
				close(c.send)
				c.send = nil
			}

			if c.process != nil {
				close(c.process)
				c.process = nil
			}
		}
	})
}

func (c *TcpConn) InitData() {

}
