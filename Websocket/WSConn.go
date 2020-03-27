package grapeWSNet

// Websocket Conn
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/3

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	ws "github.com/gorilla/websocket"
	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	utils "github.com/koangel/grapeNet/Utils"
)

type WSConn struct {
	cm.Conn

	ownerNet *WSNetwork
	WConn    *ws.Conn
	UserData interface{} // 用户对象
	LastPing time.Time

	send    chan []byte
	process chan []byte // 单独的一个数据包

	CryptKey []byte

	IsClosed int32

	readTimeout  time.Duration
	writeTimeout time.Duration
}

const (
	ReadWaitPing = 120 * time.Second
	WriteTicker  = 10 * time.Minute

	pingTickTime = 30 * time.Second

	queueCount = 2048
)

///////////////////////////////////////////////
// 新建WS
func NewWConn(wn *WSNetwork, Conn *ws.Conn, UData interface{}) *WSConn {
	NewWConn := &WSConn{
		ownerNet: wn,
		WConn:    Conn,
		UserData: UData,
		CryptKey: []byte{},

		LastPing: time.Now(),

		send:     make(chan []byte, queueCount),
		process:  make(chan []byte, queueCount),
		IsClosed: 0,

		writeTimeout: WriteTicker,
		readTimeout:  ReadWaitPing,
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
	ws.DefaultDialer.HandshakeTimeout = 120 * time.Second
	wsHeader := http.Header{}
	wsHeader.Set("Origin", sOrigin)
	wsHeader.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36")
	ws, _, derr := ws.DefaultDialer.Dial(fmt.Sprintf("ws://%v%v", addr, wn.wsPath), wsHeader)
	if derr != nil {
		logger.ERROR(derr.Error())
		err = derr
		return
	}

	conn = &WSConn{
		ownerNet: wn,
		UserData: UData,
		WConn:    ws,
		LastPing: time.Now(),
		IsClosed: 0,

		CryptKey: []byte("e63b58801d951ff2435d0a6242a44b6e34062233"),

		send:    make(chan []byte, queueCount),
		process: make(chan []byte, queueCount),

		writeTimeout: WriteTicker,
		readTimeout:  ReadWaitPing,
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
func (c *WSConn) SetReadTimeout(t time.Duration) {
	c.readTimeout = t
}

func (c *WSConn) SetWriteTimeout(t time.Duration) {
	c.writeTimeout = t
}

func (c *WSConn) startProc() {
	go c.writePump()
	go c.recvPump()
}

func (c *WSConn) handlerPump() {
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
func (c *WSConn) recvPump() {
	defer func() {
		if p := recover(); p != nil {
			stacks := utils.PanicTrace(4)
			panic := fmt.Sprintf("recover panics: %v call:%v", p, string(stacks))
			logger.ERROR(panic)

			if c.ownerNet.Panic != nil {
				c.ownerNet.Panic(c, panic)
			}
		}

		logger.INFO("%v recv Pump defer done!!!", c.SessionId)
		c.Cancel() // 结束
		c.Wg.Wait()
		c.Close()                             // 关闭SOCKET
		c.ownerNet.RemoveSession(c.SessionId) // 删除
	}()

	c.WConn.SetReadLimit(65536 * 4)
	c.WConn.SetPingHandler(func(string) error {
		c.Send([]byte{0xf1, ws.PongMessage})
		c.WConn.SetReadDeadline(time.Now().Add(c.readTimeout))
		return nil
	})
	c.WConn.SetPongHandler(func(string) error {
		c.WConn.SetReadDeadline(time.Now().Add(c.readTimeout))
		return nil
	})

	for {
		c.WConn.SetReadDeadline(time.Now().Add(c.readTimeout))
		wType, wmsg, err := c.WConn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway) {
				logger.ERROR("Session %v Recv Error:%v", c.SessionId, err)
				return
			}

			logger.ERROR("Session %v Recv Error:%v", c.SessionId, err)
			return
		}

		if wType == -1 {
			continue // 不需要处理这是个错误或PING
		}

		if wType != c.ownerNet.MsgType {
			logger.ERROR("Message Type Not Allowed:%v...", wType)
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		if c.ownerNet.OnHandler != nil {
			item := c.ownerNet.Decrypt(wmsg, c.CryptKey)
			if len(item) > 1 {
				c.ownerNet.OnHandler(c, item)
			}
		}
	}
}

func (c *WSConn) writePump() {
	c.Wg.Add(1)
	ticker := time.NewTicker(pingTickTime)
	defer func() {
		if p := recover(); p != nil {
			stacks := utils.PanicTrace(4)
			panic := fmt.Sprintf("writePump panics: %v call:%v", p, string(stacks))
			logger.ERROR(panic)

			if c.ownerNet.Panic != nil {
				c.ownerNet.Panic(c, panic)
			}
		}

		c.Wg.Done()
		logger.INFO("%v write Pump defer done!!!", c.SessionId)
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

			c.WConn.SetWriteDeadline(time.Now().Add(WriteTicker))

			if len(bData) == 2 && bData[0] == 0xf1 {
				if err := c.WConn.WriteMessage(int(bData[1]), nil); err != nil {
					logger.INFO("writePump ticker error,%v!!!", err)
					return // 在SELECT中必须使用RETUN，如果使用BREAK代表跳出SELECT，毫无意义
				}
			} else {
				if err := c.WConn.WriteMessage(c.ownerNet.MsgType, bData); err != nil {
					logger.ERROR("write Pump error:%v !!!", err)
					return
				}
			}

			break
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			if atomic.LoadInt32(&c.IsClosed) == 1 {
				return
			}

			c.WConn.SetWriteDeadline(time.Now().Add(WriteTicker))
			if err := c.WConn.WriteMessage(ws.PingMessage, nil); err != nil {
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

	select {
	case c.send <- c.ownerNet.Encrypt(data, c.CryptKey):
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

	pack, err := c.ownerNet.Package(val)
	if err != nil {
		logger.ERRORV(err)
		return -1
	}

	return c.Send(pack)
}

func (c *WSConn) SendDirect(data []byte) int {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		return -1
	}

	encode := c.ownerNet.Encrypt(data, c.CryptKey)
	err := c.WConn.WriteMessage(c.ownerNet.MsgType, encode)
	if err != nil {
		logger.ERRORV(err)
		return -1
	}
	return len(encode)
}

func (c *WSConn) SendPakDirect(val interface{}) int {
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

func (c *WSConn) Close() {
	c.Once.Do(func() {
		if c.WConn == nil {
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 0 {
			atomic.StoreInt32(&c.IsClosed, 1)

			c.ownerNet.OnClose(c)

			c.WConn.Close() // 关闭连接
		}
	})
}

func (c *WSConn) RemoveData() {
	if atomic.LoadInt32(&c.IsClosed) == 1 {
		c.WConn = nil

		logger.INFO("%v cleanup data...", c.SessionId)

		close(c.send)
		c.send = nil
		close(c.process)
		c.process = nil
	}
}

func (c *WSConn) InitData() {

}
