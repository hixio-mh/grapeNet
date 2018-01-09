// Websocket Network
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/3
package grapeWSNet

import (
	"net/http"

	"fmt"

	"github.com/gorilla/websocket"
	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
)

const (
	BinaryMsg = websocket.BinaryMessage
	TextMsg   = websocket.TextMessage
)

type WSNetwork struct {
	Origin    string
	address   string // 监听地址
	wsPath    string
	ChkOrigin bool
	upgrader  websocket.Upgrader
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
	OnUpgrade func(req *http.Request) bool

	// 打包以及加密行为
	Package   func(val interface{}) (data []byte, err error)
	Unpackage func(conn *WSConn, spak *stream.BufferIO) (data [][]byte, err error)

	// 输出panic数据
	Panic func(conn *WSConn, src string)

	Encrypt func(data []byte) []byte
	Decrypt func(data []byte) []byte

	HttpHome func(w http.ResponseWriter, r *http.Request)

	MsgType int
}

//////////////////////////////////////
// 新建函数
func NetEmptyWS(Origin, wPath string) *WSNetwork {
	NewWC := &WSNetwork{
		NetCM:     cm.NewCM(),
		Origin:    Origin,
		wsPath:    wPath,
		ChkOrigin: false,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  40960,
			WriteBufferSize: 40960,
		},
		CreateUserData: defaultCreateUserData,
		Package:        defaultBytePacker,
		Unpackage:      defaultByteData,

		Panic: defaultPanic,

		OnAccept:    defaultOnAccept,
		OnHandler:   nil,
		OnUpgrade:   nil,
		OnClose:     defaultOnClose,
		OnConnected: defaultOnConnected,

		Encrypt: defaultEncrypt,
		Decrypt: defaultDecrypt,

		MsgType: BinaryMsg,
	}

	if len(Origin) > 0 {
		NewWC.ChkOrigin = true
	}

	return NewWC
}

func NewWebsocket(addr, Origin, wPath string) *WSNetwork {
	NewWC := &WSNetwork{
		address:   addr,
		NetCM:     cm.NewCM(),
		Origin:    Origin,
		wsPath:    wPath,
		ChkOrigin: false,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  40960,
			WriteBufferSize: 40960,
		},
		CreateUserData: defaultCreateUserData,
		Package:        defaultBytePacker,
		Unpackage:      defaultByteData,

		Panic: defaultPanic,

		OnAccept:    defaultOnAccept,
		OnHandler:   nil,
		OnUpgrade:   nil,
		OnClose:     defaultOnClose,
		OnConnected: defaultOnConnected,

		Encrypt: defaultEncrypt,
		Decrypt: defaultDecrypt,

		MsgType: BinaryMsg,
	}

	if len(Origin) > 0 {
		NewWC.ChkOrigin = true
	}

	return NewWC
}

//////////////////////////////////////
// 成员函数
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

func (wn *WSNetwork) serveWs(w http.ResponseWriter, r *http.Request) {
	wn.upgrader.CheckOrigin = func(rr *http.Request) bool {
		if wn.ChkOrigin == false {
			return true
		}

		lOrigin := rr.Header.Get("Origin")
		if wn.ChkOrigin && lOrigin == wn.Origin {
			return true
		}

		return false
	}

	conn, err := wn.upgrader.Upgrade(w, r, nil) // 升级协议
	if err != nil {
		logger.ERROR(err.Error())
		return
	}

	if wn.OnUpgrade != nil && wn.OnUpgrade(r) == false {
		logger.ERROR("upgrade websocket faild...")
		conn.Close() // 断开
		return
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

func (wn *WSNetwork) defaultHome(w http.ResponseWriter, r *http.Request) {
	if wn.HttpHome != nil {
		wn.HttpHome(w, r)
	}
}

func (wn *WSNetwork) listen() error {
	if len(wn.address) == 0 {
		return fmt.Errorf("listener is nil...")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		wn.defaultHome(w, r)
	})
	http.HandleFunc(wn.wsPath, func(w http.ResponseWriter, r *http.Request) {
		wn.serveWs(w, r)
	})
	logger.INFO("Server Start On:%v", wn.address)
	err := http.ListenAndServe(wn.address, nil)
	if err != nil {
		logger.ERROR("ListenAndServe: ", err)
		return err
	}

	return nil
}
