// Websocket Conn
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/3
package grapeWSNet

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	ws "github.com/gorilla/websocket"
	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"

	stream "github.com/koangel/grapeNet/Stream"
)

type WSConn struct {
	cm.Conn

	ownerNet *WSNetwork
	WConn    *ws.Conn
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

///////////////////////////////////////////////
// 新建WS
func NewWConn(wn *WSNetwork, Conn *ws.Conn, UData interface{}) *WSConn {
	NewWConn := &WSConn{
		ownerNet: wn,
		WConn:    Conn,
		UserData: UData,

		LastPing: time.Now(),

		send:     make(chan []byte, 1024),
		process:  make(chan *stream.BufferIO, 1024),
		IsClosed: 0,
	}

	NewWConn.Ctx, NewWConn.Cancel = context.WithCancel(context.Background())
	NewWConn.Once = new(sync.Once)
	NewWConn.Wg = new(sync.WaitGroup)
	NewWConn.SessionId = cm.CreateUUID(3)
	NewWConn.Type = cm.ESERVER_TYPE

	return NewWConn
}

func NewDial(wn *WSNetwork, addr, sOrigin string, UData interface{}) (conn *WSConn, err error) {
	conn = nil
	err = errors.New("unknow error.")
	ws.DefaultDialer.HandshakeTimeout = 60 * time.Second
	hHeader := http.Header{}
	hHeader.Set("Origin", sOrigin)
	hHeader.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36")
	ws, _, derr := ws.DefaultDialer.Dial(addr, hHeader)
	if err != nil {
		logger.ERROR(err.Error())
		err = derr
		return
	}

	conn = &WSConn{
		ownerNet: wn,
		UserData: UData,
		WConn:    ws,
		LastPing: time.Now(),
		IsClosed: 0,

		send:    make(chan []byte, 1024),
		process: make(chan *stream.BufferIO, 1024),
	}

	err = nil
	conn.Ctx, conn.Cancel = context.WithCancel(context.Background())
	conn.Once = new(sync.Once)
	conn.Wg = new(sync.WaitGroup)
	conn.SessionId = cm.CreateUUID(4)
	conn.Type = cm.ECLIENT_TYPE

	return
}

//////////////////////////////////////////////
// 成员函数
func (c *WSConn) startProc() {
	go c.writePump()
	go c.recvPump()
}

func (c *WSConn) recvPump() {
	defer func() {
		if p := recover(); p != nil {
			logger.ERROR("recover panics: %v", p)
		}

		c.Cancel() // 结束
		c.Wg.Wait()
		c.Close()                             // 关闭SOCKET
		c.ownerNet.RemoveSession(c.SessionId) // 删除
	}()

	c.WConn.SetReadLimit(65536)
	c.WConn.SetReadDeadline(time.Now().Add(ReadWaitPing))

	for {
		wType, wmsg, err := c.WConn.ReadMessage()
		if err != nil {
			logger.ERROR("Session %v Recv Error:%v", c.SessionId, err)
			return
		}

		if wType != ws.BinaryMessage {
			logger.ERROR("Session %v Recv Error Type Error:%v", c.SessionId, wType)
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		c.WConn.SetReadDeadline(time.Now().Add(ReadWaitPing))
		c.mainStream.WriteAuto(wmsg)

		upak := c.ownerNet.Unpackage(c, &c.mainStream) // 调用解压行为
		for _, v := range upak {
			c.ownerNet.OnHandler(c, v)
		}
	}
}

func (c *WSConn) writePump() {
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

			c.WConn.SetWriteDeadline(time.Now().Add(60 * time.Second))
			if err := c.WConn.WriteMessage(ws.BinaryMessage, bData); err != nil {
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

			c.WConn.SetWriteDeadline(time.Now().Add(60 * time.Second))
			if err := c.WConn.WriteMessage(ws.PingMessage, []byte{}); err != nil {
				logger.INFO("writePump ticker error,%v!!!", err)
				return // 在SELECT中必须使用RETUN，如果使用BREAK代表跳出SELECT，毫无意义
			}
			break
		}

	}
}

func (c *WSConn) Send(data []byte) int {
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

func (c *WSConn) SendPak(val interface{}) int {
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

func (c *WSConn) Close() {
	c.Once.Do(func() {
		if atomic.LoadInt32(&c.IsClosed) == 0 {
			atomic.StoreInt32(&c.IsClosed, 1)

			c.ownerNet.OnClose(c)

			c.WConn.Close() // 关闭连接
		}
	})
}

func (c *WSConn) RemoveData() {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		close(c.send)
		c.send = nil
		close(c.process)
		c.process = nil
	}
}

func (c *WSConn) InitData() {

}
