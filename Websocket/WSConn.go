package grapeWSNet

// Websocket Conn
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
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
	process chan []byte

	CryptKey []byte

	IsClosed int32

	readTimeout  time.Duration
	writeTimeout time.Duration

	RMData sync.Once

	connMux sync.RWMutex

	sendMux   sync.Mutex
	closeSock sync.Once

	remoteAddr string
}

const (
	ReadWaitPing = 120 * time.Second
	WriteTicker  = 2 * time.Minute

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
	NewWConn.remoteAddr = Conn.RemoteAddr().String()

	return NewWConn
}

func NewDial(wn *WSNetwork, addr, sOrigin string, UData interface{}) (conn *WSConn, err error) {
	conn = nil
	err = errors.New("unknow error.")
	wsHeader := http.Header{}
	wsHeader.Set("Origin", sOrigin)
	wsHeader.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36")
	tctx, _ := context.WithTimeout(context.Background(), 45*time.Second) // 连接45秒超时
	ws, _, derr := ws.DefaultDialer.DialContext(tctx, fmt.Sprintf("ws://%v%v", addr, wn.wsPath), wsHeader)
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
	conn.remoteAddr = ws.RemoteAddr().String()

	return
}

//////////////////////////////////////////////
// 成员函数
func (c *WSConn) SetUserData(user interface{}) {
	c.UserData = user
}

func (c *WSConn) GetUserData() interface{} {
	return c.UserData
}

func (c *WSConn) GetNetConn() net.Conn {
	c.connMux.RLock()
	defer c.connMux.RUnlock()

	rc := c.WConn
	return rc.UnderlyingConn()
}

func (c *WSConn) RemoteAddr() string {
	c.connMux.RLock()
	defer c.connMux.RUnlock()

	return c.WConn.RemoteAddr().String()
}

func (c *WSConn) GetConn() *ws.Conn {
	c.connMux.RLock()
	defer c.connMux.RUnlock()

	rc := c.WConn
	return rc
}

func (c *WSConn) SetReadTimeout(t time.Duration) {
	c.readTimeout = t
}

func (c *WSConn) SetWriteTimeout(t time.Duration) {
	c.writeTimeout = t
}

func (c *WSConn) startProc() {
	go c.writePump()
	go c.recvPump()
	if HandlerProc > 0 {
		go c.handlerPump()
	}
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
		logger.INFO("%v handler Pump defer done!!!", c.remoteAddr)
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

func (c *WSConn) readMessage() (messageType int, p []byte, err error) {
	var r io.Reader
	wconn := c.GetConn()
	messageType, r, err = wconn.NextReader()
	if err != nil {
		return messageType, nil, err
	}

	buff := bytes.NewBuffer(make([]byte, 0, 1024))
	rBuff := make([]byte, 1024)
	for {
		rn, verr := r.Read(rBuff)
		if verr != nil || rn == 0 {
			break
		}

		buff.Write(rBuff[:rn])
	}

	temp := buff.Bytes()
	length := len(temp)
	var body []byte
	//are we wasting more than 10% space?
	if cap(temp) > (length + length/10) {
		body = make([]byte, length)
		copy(body, temp)
	} else {
		body = temp
	}
	return messageType, body, err
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

		logger.INFO("%v %v recv Pump defer done!!!", c.remoteAddr, c.SessionId)
		c.Cancel() // 结束
		c.Wg.Wait()
		c.Close()                             // 关闭SOCKET
		c.ownerNet.RemoveSession(c.SessionId) // 删除
	}()

	c.WConn.SetReadLimit(65536 * 4)
	c.WConn.SetPingHandler(func(string) error {
		if c.WConn != nil {
			c.Send([]byte{0xf1, ws.PongMessage})
			c.WConn.SetReadDeadline(time.Now().Add(c.readTimeout))
		}
		return nil
	})
	c.WConn.SetPongHandler(func(string) error {
		if c.WConn != nil {
			c.WConn.SetReadDeadline(time.Now().Add(c.readTimeout))
		}
		return nil
	})

	for {
		WConn := c.GetConn()
		if WConn == nil {
			logger.ERROR("Session %v %v Close WEBSOCKET", c.remoteAddr, c.SessionId)
			return
		}

		WConn.SetReadDeadline(time.Now().Add(c.readTimeout))
		wType, wmsg, err := WConn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway) {
				logger.ERROR("Session %v %v Recv Error:%v", c.remoteAddr, c.SessionId, err)
				return
			}

			logger.ERROR("Session %v %v Recv Error:%v", c.remoteAddr, c.SessionId, err)
			return
		}

		if wType == -1 {
			continue // 不需要处理这是个错误或PING
		}

		if wType != c.ownerNet.MsgType {
			logger.ERROR("Message Type Not Allowed:%v %v...", c.remoteAddr, wType)
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		if c.ownerNet.OnHandler != nil {
			item := c.ownerNet.Decrypt(wmsg, c.CryptKey)
			if len(item) > 1 {
				if HandlerProc <= 0 {
					if c.ownerNet.OnHandler != nil {
						c.ownerNet.OnHandler(c, item)
					}
				} else {
					if c.process != nil {
						c.process <- item
					}
				}
			}
		}
	}
}

func (c *WSConn) writeFrames(messageType int, data []byte) error {
	c.sendMux.Lock()
	defer c.sendMux.Unlock()

	WConn := c.GetConn()
	if atomic.LoadInt32(&c.IsClosed) == 1 || WConn == nil {
		return errors.New("conn is closed...")
	}

	WConn.SetWriteDeadline(time.Now().Add(WriteTicker))

	wr, err := c.WConn.NextWriter(messageType)
	if err != nil {
		logger.ERRORV(err)
		return err
	}

	_, err = wr.Write(data)
	if err != nil {
		logger.ERRORV(err)
		return err
	}

	return wr.Close()
}

// 此处保证永远只有一个在发送
func (c *WSConn) writeLockMsg(messageType int, data []byte) error {
	c.sendMux.Lock()
	defer c.sendMux.Unlock()

	WConn := c.GetConn()
	if atomic.LoadInt32(&c.IsClosed) == 1 || WConn == nil {
		return errors.New("conn is closed...")
	}

	WConn.SetWriteDeadline(time.Now().Add(WriteTicker))
	err := WConn.WriteMessage(messageType, data)
	if err != nil {
		logger.ERRORV(err)
		return err
	}

	return nil
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
		logger.INFO("%v %v write Pump defer done!!!", c.remoteAddr, c.SessionId)
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

			c.WConn.SetWriteDeadline(time.Now().Add(WriteTicker))
			if len(bData) == 2 && bData[0] == 0xf1 {
				if err := c.writeLockMsg(int(bData[1]), nil); err != nil {
					logger.INFO("%v writePump ticker error,%v!!!", c.remoteAddr, err)
					return
				}
			} else {
				if err := c.writeLockMsg(c.ownerNet.MsgType, bData); err != nil {
					logger.ERROR("%v write Pump error:%v !!!", c.remoteAddr, err)
					return
				}
			}
			break
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			WConn := c.GetConn()
			if atomic.LoadInt32(&c.IsClosed) == 1 || WConn == nil {
				return
			}

			c.WConn.SetWriteDeadline(time.Now().Add(WriteTicker))
			if err := c.writeLockMsg(ws.PingMessage, nil); err != nil {
				logger.INFO("%v writePump ticker error,%v!!!", c.remoteAddr, err)
				return
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
		logger.ERROR("%v Package Func Error,Can't Send...", c.remoteAddr)
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
	c.WConn.SetWriteDeadline(time.Now().Add(WriteTicker))
	err := c.writeLockMsg(c.ownerNet.MsgType, encode)
	if err != nil {
		logger.ERRORV(err)
		c.Close()
		return -1
	}
	return len(encode)
}

func (c *WSConn) SendPakDirect(val interface{}) int {
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

func (c *WSConn) CloseSocket() {
	c.closeSock.Do(func() {
		c.Cancel()
		atomic.StoreInt32(&c.IsClosed, 1)
		c.WConn.Close() // 关闭连接
	})
}

func (c *WSConn) Close() {
	c.Once.Do(func() {
		WConn := c.GetConn()
		if WConn == nil {
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 0 {
			c.ownerNet.OnClose(c)

			c.CloseSocket()
		}
	})
}

func (c *WSConn) RemoveData() {
	c.RMData.Do(func() {
		defer func() {
			if p := recover(); p != nil {
				stacks := utils.PanicTrace(4)
				panic := fmt.Sprintf("RemoveData panics: %v call:%v", p, string(stacks))
				logger.ERROR(panic)

				if c.ownerNet.Panic != nil {
					c.ownerNet.Panic(c, panic)
				}
			}
		}()

		if atomic.LoadInt32(&c.IsClosed) == 1 {
			// 连接信息
			c.connMux.Lock()
			c.WConn = nil
			c.connMux.Unlock()

			logger.INFO("%v %v cleanup data...", c.remoteAddr, c.SessionId)

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

func (c *WSConn) InitData() {

}
