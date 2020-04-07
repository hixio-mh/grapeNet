package kcpNet

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xtaci/kcp-go"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"

	utils "github.com/koangel/grapeNet/Utils"
)

type KcpConn struct {
	cm.Conn
	ownerNet *KcpNetwork
	TConn    net.Conn    // tcp连接
	UserData interface{} // 用户对象
	LastPing time.Time

	send    chan []byte
	process chan []byte // 单独的一个数据包

	CryptKey []byte

	IsClosed int32

	writeTime int
	readTime  int

	removeOnce sync.Once
	closeSock  sync.Once

	remoteAddr string
}

const (
	queueCount = 2048
)

const (
	RMStream   = iota // 使用流算法读写数据包
	RMReadFull        // 使用固定算法读写数据包 固定算法 包头4字节为固定长度，其余位置为数据包payload
)

//////////////////////////////////////
// 创建新连接

func EmptyConn(ctype int) *KcpConn {
	newConn := &KcpConn{
		LastPing:  time.Now(),
		IsClosed:  0,
		writeTime: 120,
		readTime:  120,
		CryptKey:  []byte{},

		send:    make(chan []byte, queueCount),
		process: make(chan []byte, queueCount),
	}

	newConn.Ctx, newConn.Cancel = context.WithCancel(context.Background())
	newConn.Once = new(sync.Once)
	newConn.Wg = new(sync.WaitGroup)
	newConn.SessionId = cm.CreateUUID(ctype)
	newConn.Type = ctype

	return newConn
}

func NewConn(tn *KcpNetwork, conn *kcp.UDPSession, UData interface{}) *KcpConn {
	newConn := EmptyConn(cm.ESERVER_TYPE)

	newConn.writeTime = tn.KcpConf.Writetimeout
	newConn.readTime = tn.KcpConf.Readtimeout

	newConn.TConn = conn
	newConn.ownerNet = tn
	newConn.UserData = UData

	newConn.remoteAddr = conn.RemoteAddr().String()

	return newConn
}

func NewDial(tn *KcpNetwork, addr string, UData interface{}) (conn *KcpConn, err error) {
	err = nil
	conn = nil

	config := tn.KcpConf

	dconn, derr := kcp.DialWithOptions(addr, tn.KcpBlock, tn.KcpConf.DataShard, tn.KcpConf.ParityShard)
	if derr != nil {
		logger.ERRORV(derr)
		err = derr
		return
	}

	dconn.SetStreamMode(true)
	dconn.SetWriteDelay(false)
	dconn.SetNoDelay(config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
	dconn.SetMtu(config.MTU)
	dconn.SetWindowSize(config.SndWnd, config.RcvWnd)
	dconn.SetACKNoDelay(config.AckNodelay)

	conn = EmptyConn(cm.ECLIENT_TYPE)

	conn.writeTime = tn.KcpConf.Writetimeout
	conn.readTime = tn.KcpConf.Readtimeout

	conn.ownerNet = tn
	conn.TConn = dconn
	conn.UserData = UData
	conn.remoteAddr = dconn.RemoteAddr().String()

	return
}

//////////////////////////////////////////////
// 成员函数
func (c *KcpConn) SetUserData(user interface{}) {
	c.UserData = user
}

func (c *KcpConn) GetUserData() interface{} {
	return c.UserData
}

func (c *KcpConn) GetNetConn() net.Conn {
	return c.TConn
}

func (c *KcpConn) RemoteAddr() string {
	return c.remoteAddr
}

func (c *KcpConn) startProc() {
	go c.writePump()
	if c.ownerNet.RecvMode == RMStream {
		go c.recvPump()
	} else {
		go c.recvPumpFull()
	}

	if HandlerProc > 0 {
		go c.handlerPump()
	}
}

func (c *KcpConn) handlerPump() {
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
		logger.INFOV("addr:", c.remoteAddr, ",", c.SessionId, " handler Pump defer done!!!")
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
					c.ownerNet.OnHandler(c, item)
				}
			}
		}
	}
}

func (c *KcpConn) recvPump() {
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

	var buffer []byte = make([]byte, 65535)
	var lStream stream.BufferIO

	for {
		c.TConn.SetReadDeadline(time.Now().Add(time.Duration(c.readTime) * time.Second))

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

		// 拆包
		if lStream.Write(buffer, rn) == -1 {
			logger.ERROR("Session %v %v Recv Error:Packet size is too big...", c.remoteAddr, c.SessionId)
			return
		}

		upak, err := c.ownerNet.Unpackage(c, &lStream) // 调用解压行为
		if err == nil {
			for _, v := range upak {
				// 心跳包
				if c.ownerNet.SendPong(c, v) {
					continue
				}

				if HandlerProc <= 0 {
					if c.ownerNet.OnHandler != nil {
						c.ownerNet.OnHandler(c, v[4:])
					}
				} else {
					if c.process != nil {
						c.process <- v[4:]
					}
				}
			}
		}
	}
}

func (c *KcpConn) recvPumpFull() {
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
		c.TConn.SetReadDeadline(time.Now().Add(time.Duration(c.readTime) * time.Second))

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
				c.process <- payload
			}
		}
	}
}

func (c *KcpConn) writePump() {
	heartbeat := time.NewTicker(10 * time.Second)
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

			c.TConn.SetWriteDeadline(time.Now().Add(time.Duration(c.writeTime) * time.Second))
			if _, err := c.TConn.Write(bData); err != nil {
				logger.ERROR("%v write Pump error:%v !!!", c.remoteAddr, err)
				c.CloseSocket()
				return
			}
			break
		case <-heartbeat.C:
			if atomic.LoadInt32(&c.IsClosed) == 1 {
				return
			}

			c.TConn.SetWriteDeadline(time.Now().Add(time.Duration(c.writeTime) * time.Second))
			c.ownerNet.SendPing(c) // 发送心跳
			break
		}
	}
}

func (c *KcpConn) Send(data []byte) int {
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

func (c *KcpConn) SendPak(val interface{}) int {
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

func (c *KcpConn) SendDirect(data []byte) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	encode, err := stream.PackerOnce(data, c.ownerNet.Encrypt, c.CryptKey)
	if err != nil {
		return -1
	}
	c.TConn.SetWriteDeadline(time.Now().Add(time.Duration(c.writeTime) * time.Second))
	wn, err := c.TConn.Write(encode)
	if err != nil {
		logger.ERRORV(err)
		return -1
	}
	return wn
}

func (c *KcpConn) SendPakDirect(val interface{}) int {
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

func (c *KcpConn) CloseSocket() {
	c.closeSock.Do(func() {
		atomic.StoreInt32(&c.IsClosed, 1)
		c.TConn.Close() // 关闭连接
	})
}

func (c *KcpConn) Close() {
	c.Once.Do(func() {
		// 都没连上怎么关
		if c.TConn == nil {
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 0 {
			c.ownerNet.OnClose(c)

			c.TConn.Close() // 关闭连接
			c.CloseSocket()
		}
	})
}

func (c *KcpConn) RemoveData() {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		close(c.send)
		c.send = nil
		close(c.process)
		c.process = nil
	}
}

func (c *KcpConn) InitData() {

}
