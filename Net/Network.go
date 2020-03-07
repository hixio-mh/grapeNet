// 网络层可以使用多种协议
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/10

package grapeNet

import (
	"net"

	"fmt"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
)

type TCPNetwork struct {
	listener net.Listener

	NetCM *cm.ConnManager

	/// 所有的callBack函数
	// 创建用户DATA
	CreateUserData func() interface{}

	// 通知连接
	OnAccept func(conn *TcpConn)
	// 数据包进入
	OnHandler func(conn *TcpConn, ownerPak []byte)
	// 连接关闭
	OnClose func(conn *TcpConn)
	// 连接成功
	OnConnected func(conn *TcpConn)

	MainProc func() // 简易主处理函数

	// 打包以及加密行为
	Package   func(val interface{}) (data []byte, err error)
	Unpackage func(conn *TcpConn, spak *stream.BufferIO) (data [][]byte, err error)

	// 输出panic数据
	Panic func(conn *TcpConn, src string)

	Encrypt func(data, key []byte) []byte
	Decrypt func(data, key []byte) []byte

	// ping,pong CALL
	SendPing func(conn *TcpConn)
	SendPong func(conn *TcpConn, ping []byte) bool
}

var (
	HandlerProc = 2
)

/////////////////////////////
// 创建网络服务器
func NewTcpServer(addr string) (tcp *TCPNetwork, err error) {

	if HandlerProc <= 1 {
		HandlerProc = 1
	}

	tcp = &TCPNetwork{
		listener: nil,
		NetCM:    cm.NewCM(),

		CreateUserData: defaultCreateUserData,
		Package:        defaultBytePacker, // bson转换或打包
		Unpackage:      defaultByteData,

		OnAccept:    defaultOnAccept,
		OnHandler:   nil,
		OnClose:     defaultOnClose,
		OnConnected: defaultOnConnected,

		MainProc: defaultMainProc,

		Encrypt: defaultEncrypt,
		Decrypt: defaultDecrypt,

		Panic: defaultPanic,

		SendPing: defaultPing,
		SendPong: defalutPong,
	}

	err = tcp.listen(addr)
	if err != nil {
		tcp = nil
	}
	return
}

func NewEmptyTcp() *TCPNetwork {
	if HandlerProc <= 1 {
		HandlerProc = 1
	}

	return &TCPNetwork{
		listener: nil,
		NetCM:    cm.NewCM(),

		CreateUserData: defaultCreateUserData,
		Package:        defaultBytePacker, // bson转换或打包
		Unpackage:      defaultByteData,

		OnAccept:    defaultOnAccept,
		OnHandler:   nil,
		OnClose:     defaultOnClose,
		OnConnected: defaultOnConnected,

		MainProc: defaultMainProc,

		Encrypt: defaultEncrypt,
		Decrypt: defaultDecrypt,

		Panic: defaultPanic,

		SendPing: defaultPing,
		SendPong: defalutPong,
	}
}

////////////////////////////
// 成员函数
func (c *TCPNetwork) RemoveSession(sessionId string) {
	c.NetCM.Remove(sessionId)
}

func (c *TCPNetwork) Dial(addr string, UserData interface{}) (conn *TcpConn, err error) {
	logger.INFO("Dial To :%v", addr)
	conn, err = NewDial(c, addr, UserData)
	if err != nil {
		logger.ERROR("Dial Faild:%v", err)
		return
	}

	c.OnConnected(conn)
	c.NetCM.Register <- conn // 注册账户
	conn.startProc()

	return
}

func (c *TCPNetwork) listen(bindAddr string) error {
	if c.listener != nil {
		return fmt.Errorf("listener is nil...")
	}

	lis, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return err
	}

	logger.INFO("grapeNet listen On:[%v]", bindAddr)

	c.listener = lis
	go c.onAccept()
	return nil
}

/// 连接池的处理
func (c *TCPNetwork) onAccept() {
	defer func() {
		if p := recover(); p != nil {
			logger.ERROR("recover panics: %v", p)
		}
	}()
	// 1000次错误 跳出去
	for failures := 0; failures < 1000; {
		conn, listenErr := c.listener.Accept()
		if listenErr != nil {
			logger.ERROR("accept error:%v", listenErr)
			failures++
			continue
		}

		logger.INFO("New Connection:%v，Accept.", conn.RemoteAddr())
		var client = NewConn(c, conn, c.CreateUserData())

		c.OnAccept(client)

		c.NetCM.Register <- client // 注册一个全局对象

		client.startProc() // 启动线程
	}
}

func (c *TCPNetwork) Runnable() {
	c.MainProc()
}
