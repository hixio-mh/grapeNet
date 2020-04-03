// Websocket v2 Network
// version 2.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2020/04/02
package grapeWSNetv2

import (
	"github.com/gobwas/ws"
	"log"
	"net"
	"net/http"
	"strings"

	"fmt"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
)

const (
	BinaryMsg = ws.OpBinary
	TextMsg   = ws.OpText
)

type WSNetwork struct {
	Origin    string
	address   string // 监听地址
	wsPath    string
	ChkOrigin bool
	upgrader  ws.Upgrader
	NetCM     *cm.ConnManager

	/// 所有的callBack函数
	// 创建用户DATA
	CreateUserData func() interface{}

	// 通知连接
	OnAccept func(conn *WSConn)
	// 数据包进入
	OnHandler func(conn *WSConn, ownerPak []byte)
	// 连接关闭
	OnClose func(conn *WSConn)
	// 连接成功
	OnConnected func(conn *WSConn)

	// 连接安全性检测 server only
	OnHeader  func(key, value string) error
	OnRequest func(url string) error

	// 打包以及加密行为
	Package   func(val interface{}) (data []byte, err error)
	Unpackage func(conn *WSConn, spak *stream.BufferIO) (data [][]byte, err error)

	// 输出panic数据
	Panic func(conn *WSConn, src string)

	Encrypt func(data, key []byte) []byte
	Decrypt func(data, key []byte) []byte

	HttpHome func(w http.ResponseWriter, r *http.Request)

	MsgType ws.OpCode
}

var (
	HandlerProc = 2
)

//////////////////////////////////////
// 新建函数
func NetEmptyWS(Origin, wPath string) *WSNetwork {
	if HandlerProc <= 1 {
		HandlerProc = 1
	}

	NewWC := &WSNetwork{
		NetCM:          cm.NewCM(),
		Origin:         Origin,
		wsPath:         wPath,
		ChkOrigin:      false,
		CreateUserData: defaultCreateUserData,
		Package:        defaultBytePacker,
		Unpackage:      defaultByteData,

		Panic: defaultPanic,

		OnAccept:    defaultOnAccept,
		OnHandler:   nil,
		OnHeader:    nil,
		OnRequest:   nil,
		OnClose:     defaultOnClose,
		OnConnected: defaultOnConnected,

		Encrypt: defaultEncrypt,
		Decrypt: defaultDecrypt,

		MsgType: BinaryMsg,
	}

	if len(Origin) > 0 {
		NewWC.ChkOrigin = true
	}

	NewWC.upgrader = ws.Upgrader{
		OnRequest: func(uri []byte) error {
			if !NewWC.CheckPath(string(uri)) {
				return ws.RejectConnectionError(
					ws.RejectionReason("request uri..."),
					ws.RejectionStatus(404),
				)
			}

			if NewWC.OnRequest != nil {
				err := NewWC.OnRequest(string(uri))
				if err != nil {
					return ws.RejectConnectionError(ws.RejectionStatus(403), ws.RejectionReason(err.Error()))
				}
			}
			return nil
		},
		OnHeader: func(key, value []byte) error {
			if NewWC.CheckOrigin(string(key), string(value)) == false {
				return ws.RejectConnectionError(
					ws.RejectionReason("bad request data..."),
					ws.RejectionStatus(403),
				)
			}

			err := NewWC.OnHeader(string(key), string(value))
			if err != nil {
				return ws.RejectConnectionError(ws.RejectionStatus(403), ws.RejectionReason(err.Error()))
			}
			return nil
		},
	}

	return NewWC
}

func NewWebsocket(addr, Origin, wPath string) *WSNetwork {
	if HandlerProc <= 1 {
		HandlerProc = 1
	}

	NewWC := &WSNetwork{
		address:        addr,
		NetCM:          cm.NewCM(),
		Origin:         Origin,
		wsPath:         wPath,
		ChkOrigin:      false,
		CreateUserData: defaultCreateUserData,
		Package:        defaultBytePacker,
		Unpackage:      defaultByteData,

		Panic: defaultPanic,

		OnAccept:    defaultOnAccept,
		OnHandler:   nil,
		OnHeader:    nil,
		OnRequest:   nil,
		OnClose:     defaultOnClose,
		OnConnected: defaultOnConnected,

		Encrypt: defaultEncrypt,
		Decrypt: defaultDecrypt,

		MsgType: BinaryMsg,
	}

	if len(Origin) > 0 {
		NewWC.ChkOrigin = true
	}

	NewWC.upgrader = ws.Upgrader{
		OnRequest: func(uri []byte) error {
			if !NewWC.CheckPath(string(uri)) {
				return ws.RejectConnectionError(
					ws.RejectionReason("request uri..."),
					ws.RejectionStatus(404),
				)
			}

			if NewWC.OnRequest != nil {
				err := NewWC.OnRequest(string(uri))
				if err != nil {
					return ws.RejectConnectionError(ws.RejectionStatus(403), ws.RejectionReason(err.Error()))
				}
			}
			return nil
		},
		OnHeader: func(key, value []byte) error {
			if NewWC.OnHeader != nil {
				if NewWC.CheckOrigin(string(key), string(value)) == false {
					return ws.RejectConnectionError(
						ws.RejectionReason("bad request data..."),
						ws.RejectionStatus(403),
					)
				}

				err := NewWC.OnHeader(string(key), string(value))
				if err != nil {
					return ws.RejectConnectionError(ws.RejectionStatus(403), ws.RejectionReason(err.Error()))
				}
			}
			return nil
		},
	}

	return NewWC
}

//////////////////////////////////////
// 成员函数
func (c *WSNetwork) CheckPath(url string) bool {
	return strings.HasSuffix(strings.ToLower(url), c.wsPath)
}

func (c *WSNetwork) CheckOrigin(key, value string) bool {
	if c.ChkOrigin == false {
		return true
	}

	if !strings.HasPrefix(strings.ToLower(string(key)), "origin") {
		return true
	}

	return value == c.Origin
}

func (c *WSNetwork) SetTextMessage() {
	c.MsgType = TextMsg
}

func (c *WSNetwork) SetBinaryMessage() {
	c.MsgType = BinaryMsg
}

func (c *WSNetwork) RemoveSession(sessionId string) {
	c.NetCM.Remove(sessionId)
}

func (c *WSNetwork) Dial(addr string) (conn *WSConn, err error) {
	logger.INFO("Dial To :%v", addr)
	conn, err = NewDial(c, addr, c.Origin, c.CreateUserData())
	if err != nil {
		logger.ERROR("Dial Faild:%v", err)
		return
	}

	c.OnConnected(conn)
	c.NetCM.Register <- conn // 注册账户
	conn.startProc()

	return
}

func (c *WSNetwork) Runnable() {
	werr := c.listen()
	if werr != nil {
		logger.ERROR("Listen WS Error:%v", werr)
		return
	}

}

func (wn *WSNetwork) defaultHome(w http.ResponseWriter, r *http.Request) {
	if wn.HttpHome != nil {
		wn.HttpHome(w, r)
	}
}

func (wn *WSNetwork) acceptor(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.ERRORV("acceptor error:", err)
			break
		}

		_, err = wn.upgrader.Upgrade(conn)
		if err != nil {
			logger.ERRORV("Upgrade error:", err)
			conn.Close() // 断开连接
			continue
		}

		newCnc := NewWConn(wn, conn, wn.CreateUserData())
		if newCnc == nil {
			conn.Close()
			logger.ERROR("Create Conn Error...")
			return
		}

		wn.OnAccept(newCnc)
		wn.NetCM.Register <- newCnc
		newCnc.startProc()
	}
}

func (wn *WSNetwork) listen() error {
	if len(wn.address) == 0 {
		return fmt.Errorf("listener is nil...")
	}

	ln, err := net.Listen("tcp", wn.address)
	if err != nil {
		log.Fatal(err)
	}

	logger.INFO("Server Start On:%v", wn.address)
	wn.acceptor(ln)
	return nil
}
