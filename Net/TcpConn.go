// 连接对象
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/10

package grapeNet

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
)

type TcpConn struct {
	cm.Conn
	ownerNet *TCPNetwork
	TConn    net.Conn    // tcp连接
	UserData interface{} // 用户对象
	LastPing time.Time

	mainStream stream.BufferIO

	send    chan []byte
	process chan *stream.BufferIO // 单独的一个数据包

	IsClosed int32
}

const (
	ReadWaitPing = 45 * time.Second
	WriteTicker  = 45 * time.Second
)

//////////////////////////////////////
// 创建新连接
func NewConn(tn *TCPNetwork, conn net.Conn, UData interface{}) *TcpConn {
	newConn := &TcpConn{
		ownerNet: tn,
		TConn:    conn,
		UserData: UData,
		LastPing: time.Now(),
		IsClosed: 0,

		send:    make(chan []byte, 1024),
		process: make(chan *stream.BufferIO, 1024),
	}

	newConn.Ctx, newConn.Cancel = context.WithCancel(context.Background())
	newConn.Once = new(sync.Once)
	newConn.Wg = new(sync.WaitGroup)
	newConn.SessionId = cm.CreateUUID(1)
	newConn.Type = cm.ESERVER_TYPE

	return newConn
}

func NewDial(tn *TCPNetwork, addr string, UData interface{}) (conn *TcpConn, err error) {
	err = nil
	conn = nil
	dconn, derr := net.DialTimeout("tcp", addr, time.Second*300)
	if err != nil {
		logger.ERROR(err.Error())
		err = derr
		return
	}

	conn = &TcpConn{
		ownerNet: tn,
		UserData: UData,
		LastPing: time.Now(),
		IsClosed: 0,

		send:    make(chan []byte, 1024),
		process: make(chan *stream.BufferIO, 1024),
	}

	conn.TConn = dconn
	conn.Ctx, conn.Cancel = context.WithCancel(context.Background())
	conn.Once = new(sync.Once)
	conn.Wg = new(sync.WaitGroup)
	conn.SessionId = cm.CreateUUID(2)
	conn.Type = cm.ECLIENT_TYPE

	return
}

//////////////////////////////////////////////
// 成员函数
func (c *TcpConn) startProc() {
	go c.writePump()
	go c.recvPump()
}

func (c *TcpConn) recvPump() {
	defer func() {
		if p := recover(); p != nil {
			logger.ERROR("recover panics: %v", p)
		}

		c.Cancel() // 结束
		c.Wg.Wait()
		c.Close()                             // 关闭SOCKET
		c.ownerNet.RemoveSession(c.SessionId) // 删除
	}()

	var buffer []byte = make([]byte, 65535)

	c.TConn.SetReadDeadline(time.Now().Add(ReadWaitPing))

	for {
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

		c.TConn.SetReadDeadline(time.Now().Add(ReadWaitPing))
		c.mainStream.Write(buffer, rn)
		for {
			buf, berr := c.mainStream.Unpack(true, c.ownerNet.Decrypt)
			if berr != nil {
				break
			}

			c.ownerNet.OnHandler(c, buf)
		}
	}
}

func (c *TcpConn) writePump() {
	c.Wg.Add(1)
	ticker := time.NewTicker(WriteTicker)
	defer func() {
		if p := recover(); p != nil {
			logger.ERROR("recover panics: %v", p)
		}

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

			c.TConn.SetWriteDeadline(time.Now().Add(60 * time.Second))
			if _, err := c.TConn.Write(bData); err != nil {
				logger.ERROR("write Pump error:%v !!!", err)
				return
			}

			break
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			if atomic.LoadInt32(&c.IsClosed) == 1 {
				return
			}
			break
		}

	}
}

func (c *TcpConn) Send(data []byte) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	encode := c.ownerNet.Encrypt(data)

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

	pack := c.ownerNet.Package(val)
	return c.Send(pack)
}

func (c *TcpConn) Close() {
	c.Once.Do(func() {
		if atomic.LoadInt32(&c.IsClosed) == 0 {
			atomic.StoreInt32(&c.IsClosed, 1)

			c.ownerNet.OnClose(c)

			c.TConn.Close() // 关闭连接
		}
	})
}

func (c *TcpConn) RemoveData() {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		close(c.send)
		c.send = nil
		close(c.process)
		c.process = nil
	}
}

func (c *TcpConn) InitData() {

}
