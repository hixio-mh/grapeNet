// Websocket Network
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/3
package grapeWSNet

import (
	"net/http"

	"github.com/gorilla/websocket"
	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
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

	MainProc func() // 简易主处理函数

	// 打包以及加密行为
	Package   func(val interface{}) []byte
	Unpackage func(conn *WSConn, spak *stream.BufferIO) [][]byte

	Encrypt func(data []byte) []byte
	Decrypt func(data []byte) []byte

	HttpHome func(w http.ResponseWriter, r *http.Request)
}

//////////////////////////////////////
// 新建函数
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
		Package:        nil,
		Unpackage:      DefaultByteData,

		OnAccept:    defaultOnAccept,
		OnHandler:   defaultOnHandler,
		OnClose:     defaultOnClose,
		OnConnected: defaultOnConnected,

		MainProc: defaultMainProc,

		Encrypt: defaultEncrypt,
		Decrypt: defaultDecrypt,
	}

	if len(Origin) > 0 {
		NewWC.ChkOrigin = true
	}

	err := NewWC.Listen()
	if err != nil {
		logger.ERROR("Listen WS Error:%v", err)
		return nil
	}

	return NewWC
}

//////////////////////////////////////
// 成员函数
func (c *WSNetwork) RemoveSession(sessionId string) {
	c.NetCM.Remove(sessionId)
}

func (wn *WSNetwork) serveWs(w http.ResponseWriter, r *http.Request) {
	wn.upgrader.CheckOrigin = func(rr *http.Request) bool {
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

func (wn *WSNetwork) Listen() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		wn.defaultHome(w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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
