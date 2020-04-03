package grapeWSNetv2

// Websocket Conn
// version 2.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gobwas/ws/wsutil"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	utils "github.com/koangel/grapeNet/Utils"
)

type WSConn struct {
	cm.Conn

	ownerNet *WSNetwork
	WConn    net.Conn
	UserData interface{} // 用户对象
	LastPing time.Time

	send chan []byte

	CryptKey []byte

	State ws.State

	IsClosed int32

	readTimeout  time.Duration
	writeTimeout time.Duration

	RMData sync.Once

	connMux sync.RWMutex
	sendMux sync.Mutex
}

const (
	ReadWaitPing = 120 * time.Second
	WriteTicker  = 2 * time.Minute

	pingTickTime = 30 * time.Second

	queueCount = 2048
)

///////////////////////////////////////////////
// 新建WS
func NewWConn(wn *WSNetwork, Conn net.Conn, UData interface{}) *WSConn {
	NewWConn := &WSConn{
		ownerNet: wn,
		State:    ws.StateServerSide,
		WConn:    Conn,
		UserData: UData,
		CryptKey: []byte{},

		LastPing: time.Now(),

		send:     make(chan []byte, queueCount),
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
	wsHeader := http.Header{}
	wsHeader.Set("Origin", sOrigin)
	wsHeader.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36")
	tctx, _ := context.WithTimeout(context.Background(), 50*time.Second) // 连接45秒超时
	ws.DefaultUpgrader.Header = ws.HandshakeHeaderHTTP(wsHeader)
	vconn, _, _, derr := ws.DefaultDialer.Dial(tctx, fmt.Sprintf("ws://%v%v", addr, wn.wsPath))
	if derr != nil {
		logger.ERROR(derr.Error())
		err = derr
		return
	}

	conn = &WSConn{
		State:    ws.StateClientSide,
		ownerNet: wn,
		UserData: UData,
		WConn:    vconn,
		LastPing: time.Now(),
		IsClosed: 0,

		CryptKey: []byte("e63b58801d951ff2435d0a6242a44b6e34062233"),

		send: make(chan []byte, queueCount),

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
func (c *WSConn) SetUserData(user interface{}) {
	c.UserData = user
}

func (c *WSConn) GetUserData() interface{} {
	return c.UserData
}

func (c *WSConn) GetNetConn() net.Conn {
	c.connMux.RLock()
	defer c.connMux.RUnlock()
	return c.WConn
}

func (c *WSConn) RemoteAddr() string {
	c.connMux.RLock()
	defer c.connMux.RUnlock()

	return c.WConn.RemoteAddr().String()
}

func (c *WSConn) GetConn() net.Conn {
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
}

func (c *WSConn) readData(rw io.ReadWriter) ([]byte, ws.OpCode, error) {
	controlHandler := wsutil.ControlFrameHandler(rw, c.State)
	rd := wsutil.Reader{
		Source:          rw,
		State:           c.State,
		CheckUTF8:       false,
		SkipHeaderCheck: false,
		OnIntermediate:  controlHandler,
	}
	want := c.ownerNet.MsgType | ws.OpPing
	for {
		hdr, err := rd.NextFrame()
		if err != nil {
			return nil, 0, err
		}
		if hdr.OpCode.IsControl() {
			if err := controlHandler(hdr, &rd); err != nil {
				return nil, 0, err
			}
			continue
		}
		if hdr.OpCode&want == 0 {
			if err := rd.Discard(); err != nil {
				return nil, 0, err
			}
			continue
		}

		bts, err := ioutil.ReadAll(&rd)

		return bts, hdr.OpCode, err
	}
}

func (c *WSConn) readMessage(r io.Reader) ([]wsutil.Message, error) {
	var m []wsutil.Message
	rd := wsutil.Reader{
		Source:    r,
		State:     c.State,
		CheckUTF8: false,
		OnIntermediate: func(hdr ws.Header, src io.Reader) error {
			bts, err := ioutil.ReadAll(src)
			if err != nil {
				return err
			}
			m = append(m, wsutil.Message{hdr.OpCode, bts})
			return nil
		},
	}

	h, err := rd.NextFrame()
	if err != nil {
		return m, err
	}
	var p []byte
	if h.Fin {
		// No more frames will be read. Use fixed sized buffer to read payload.
		p = make([]byte, h.Length)
		// It is not possible to receive io.EOF here because Reader does not
		// return EOF if frame payload was successfully fetched.
		// Thus we consistent here with io.Reader behavior.
		_, err = io.ReadFull(&rd, p)
	} else {
		// Frame is fragmented, thus use ioutil.ReadAll behavior.
		var buf bytes.Buffer
		_, err = buf.ReadFrom(&rd)
		p = buf.Bytes()
	}
	if err != nil {
		return m, err
	}
	return append(m, wsutil.Message{h.OpCode, p}), nil
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

	WConn := c.GetConn()
	if WConn == nil {
		logger.ERROR("Session %v Close WEBSOCKET", c.SessionId)
		return
	}

	for {
		WConn.SetReadDeadline(time.Now().Add(c.readTimeout))

		payload, opcode, err := c.readData(WConn)
		if err != nil {
			logger.ERROR("Session Read Full %v Recv Error:%v", c.SessionId, err)
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 1 {
			return
		}

		// 自动PING POING
		switch opcode {
		case ws.OpPong:
			WConn.SetWriteDeadline(time.Now().Add(WriteTicker))
			continue
		case ws.OpPing:
			WConn.SetWriteDeadline(time.Now().Add(WriteTicker))
			wsutil.WriteMessage(WConn, c.State, ws.OpPong, []byte{0})
			continue
		case ws.OpClose:
			return
		}

		if c.ownerNet.OnHandler != nil {
			item := c.ownerNet.Decrypt(payload, c.CryptKey)
			if len(item) > 1 {
				c.ownerNet.OnHandler(c, item)
			}
		}
	}
}

// 此处保证永远只有一个在发送
func (c *WSConn) writeLockMsg(messageType ws.OpCode, data []byte) error {
	// 发送端 也存在线程争抢，如果去掉，那么会导致发送数据紊乱
	// 应该是大部分ws库的通病
	c.sendMux.Lock()
	defer c.sendMux.Unlock()

	WConn := c.GetConn()
	if atomic.LoadInt32(&c.IsClosed) == 1 || WConn == nil {
		return errors.New("conn is closed...")
	}

	WConn.SetWriteDeadline(time.Now().Add(WriteTicker))

	wr := wsutil.GetWriter(WConn, c.State, messageType, len(data))
	if wr == nil {
		return fmt.Errorf("new writer error...")
	}
	defer wsutil.PutWriter(wr) // 放回去

	_, err := wr.Write(data)
	if err != nil {
		logger.ERRORV(err)
		return err
	}

	return wr.Flush()
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
			if err := c.writeLockMsg(c.ownerNet.MsgType, bData); err != nil {
				logger.ERROR("write Pump error:%v !!!", err)
				c.Close()
				return
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
			if err := c.writeLockMsg(ws.OpPing, nil); err != nil {
				logger.INFO("writePump ticker error,%v!!!", err)
				c.Close()
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
		WConn := c.GetConn()
		if WConn == nil {
			return
		}

		if atomic.LoadInt32(&c.IsClosed) == 0 {
			atomic.StoreInt32(&c.IsClosed, 1)

			c.ownerNet.OnClose(c)

			WConn.Close() // 关闭连接
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

			logger.INFO("%v cleanup data...", c.SessionId)

			if c.send != nil {
				close(c.send)
				c.send = nil
			}
		}
	})
}

func (c *WSConn) InitData() {

}
