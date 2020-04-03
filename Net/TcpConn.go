// 连接对象
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/10

package grapeNet

import (
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
	process chan []byte // 单独的一个数据包

	CryptKey []byte

	IsClosed int32

	recvMode int
}

const (
	ReadWaitPing = 65 * time.Second
	WriteTicker  = 3 * time.Minute

	queueCount    = 2048
	maxPacketSize = 6 * 1024 * 1024 // 6兆数据包最大
)

const (
	RMStream   = iota // 使用流算法读写数据包
	RMReadFull        // 使用固定算法读写数据包 固定算法 包头4字节为固定长度，其余位置为数据包payload
)

//////////////////////////////////////
// 创建新连接

func EmptyConn(ctype, mode int) *TcpConn {
	newConn := &TcpConn{
		LastPing: time.Now(),
		IsClosed: 0,

		CryptKey: []byte{},

		send: make(chan []byte, queueCount),
	}

	newConn.Ctx, newConn.Cancel = context.WithCancel(context.Background())
	newConn.Once = new(sync.Once)
	newConn.Wg = new(sync.WaitGroup)
	newConn.SessionId = cm.CreateUUID(ctype)
	newConn.Type = ctype
	newConn.recvMode = mode

	return newConn
}

func NewConn(tn *TCPNetwork, conn net.Conn, UData interface{}) *TcpConn {
	newConn := EmptyConn(cm.ESERVER_TYPE, tn.RecvMode)

	newConn.TConn = conn
	newConn.ownerNet = tn
	newConn.UserData = UData

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
		//tcpConn.SetNoDelay(false)
	}

	conn = EmptyConn(cm.ECLIENT_TYPE, tn.RecvMode)
	conn.ownerNet = tn
	conn.TConn = dconn
	conn.UserData = UData

	return
}

//////////////////////////////////////////////
// 成员函数
func (c *TcpConn) GetNetConn() net.Conn {
	return c.TConn
}

func (c *TcpConn) RemoteAddr() string {
	return c.TConn.RemoteAddr().String()
}

func (c *TcpConn) startProc() {

	go c.writePump()

	if c.recvMode == RMStream {
		go c.recvPump()
	} else {
		go c.recvPumpFull()
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
		logger.INFO("handler Pump defer done!!!")
	}()

	c.Wg.Add(1)
	for {
		select {
		case <-c.Ctx.Done():
			logger.INFO("%v session handler done...", c.SessionId)
			return
		case item := <-c.process:
			{
				if atomic.LoadInt32(&c.IsClosed) == 1 {
					return
				}

				if c.ownerNet.OnHandler != nil {
					c.ownerNet.OnHandler(c, item)
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

	}()

	var buffer = make([]byte, 65535)
	var lStream stream.BufferIO

	for {
		c.TConn.SetReadDeadline(time.Now().Add(ReadWaitPing))
		rn, err := c.TConn.Read(buffer)
		if err != nil {
			logger.ERROR("Session %v Recv Error:%v", c.SessionId, err)
			return
		}

		if rn == 0 {
			logger.ERROR("Session %v Recv Len:%v", c.SessionId, rn)
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		if c.ownerNet.OnHandler == nil {
			logger.ERROR("Handler is Null,Closed...")
			return
		}

		if lStream.Write(buffer, rn) == -1 {
			logger.ERROR("Session %v Recv Error:Packet size is too big...", c.SessionId)
			return
		}

		upak, err := c.ownerNet.Unpackage(c, &lStream) // 调用解压行为
		if err == nil {
			for _, v := range upak {
				if c.ownerNet.SendPong(c, v) {
					continue
				}

				if c.ownerNet.OnHandler != nil {
					c.ownerNet.OnHandler(c, v[4:])
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
	}()

	for {
		c.TConn.SetReadDeadline(time.Now().Add(ReadWaitPing))
		lenBytes := make([]byte, 4)
		rn, err := io.ReadFull(c.TConn, lenBytes)
		if err != nil {
			logger.ERROR("Session %v Recv Error:%v", c.SessionId, err)
			return
		}

		if rn == 0 {
			logger.ERROR("Session %v Recv Len:%v", c.SessionId, rn)
			return
		}

		headerLen := binary.LittleEndian.Uint32(lenBytes)
		payload := make([]byte, headerLen)
		rn, err = io.ReadFull(c.TConn, payload)
		if err != nil {
			logger.ERROR("Session %v Recv Error:%v", c.SessionId, err)
			return
		}

		if rn == 0 {
			logger.ERROR("Session %v Recv Len:%v", c.SessionId, rn)
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		if c.ownerNet.OnHandler == nil {
			logger.ERROR("Handler is Null,Closed...")
			return
		}

		payload = c.ownerNet.Decrypt(payload, c.CryptKey)
		if c.ownerNet.SendPong(c, payload) {
			continue
		}

		if c.ownerNet.OnHandler != nil {
			c.ownerNet.OnHandler(c, payload)
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
		logger.INFO("write Pump defer done!!!")
	}()

	for {
		select {
		case <-c.Ctx.Done():
			logger.INFO("%v session write done...", c.SessionId)
			return
		case bData, ok := <-c.send:
			if !ok {
				return
			}

			if atomic.LoadInt32(&c.IsClosed) == 1 {
				return
			}

			c.TConn.SetWriteDeadline(time.Now().Add(WriteTicker))
			if _, err := c.TConn.Write(bData); err != nil {
				logger.ERROR("write Pump error:%v !!!", err)
				c.Close()
			}
			break
		case <-heartbeat.C:
			if atomic.LoadInt32(&c.IsClosed) == 1 {
				return
			}

			c.TConn.SetWriteDeadline(time.Now().Add(WriteTicker))
			c.Send([]byte("ping")) // 发送心跳
			break
		}
	}
}

func (c *TcpConn) Send(data []byte) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	encode, err := stream.PackerOnce(data, c.ownerNet.Encrypt, c.CryptKey)
	if err != nil {
		return -1
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
		logger.ERROR("Package Func Error,Can't Send...")
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

	encode, err := stream.PackerOnce(data, c.ownerNet.Encrypt, c.CryptKey)
	if err != nil {
		return -1
	}

	c.TConn.SetWriteDeadline(time.Now().Add(WriteTicker))
	wn, err := c.TConn.Write(encode)
	if err != nil {
		logger.ERRORV(err)
		return -1
	}
	return wn
}

func (c *TcpConn) SendPakDirect(val interface{}) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	if c.ownerNet.Package == nil {
		logger.ERROR("Package Func Error,Can't Send...")
		return -1
	}

	pack, err := c.ownerNet.Package(val)
	if err != nil {
		return -1
	}

	return c.SendDirect(pack)
}

func (c *TcpConn) Close() {
	c.Once.Do(func() {
		// 都没连上怎么关
		if c.TConn == nil {
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 0 {
			atomic.StoreInt32(&c.IsClosed, 1)

			c.ownerNet.OnClose(c)

			c.TConn.Close() // 关闭连接
		}
	})
}

func (c *TcpConn) RemoveData() {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		if c.send != nil {
			close(c.send)
			c.send = nil
		}
	}
}

func (c *TcpConn) InitData() {

}
