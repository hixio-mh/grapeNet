// 网络层可以使用多种协议
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/10

package grapeNet

import (
	utils "github.com/koangel/grapeNet/Utils"
	"log"
	"net"
	//"time"

	"fmt"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
)

type TCPNetwork struct {
	listener *net.TCPListener

	NetCM *cm.ConnManager

	RecvMode  int
	SendRetry int

	UseHeaderLen bool // 是否自动附加头部4字节(RMFULL模式下强制附加)

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
	HandlerProc = 1
)

/////////////////////////////
// 创建网络服务器
func NewTcpServer(mode int, addr string) (tcp *TCPNetwork, err error) {

	if HandlerProc <= 0 {
		HandlerProc = 0
	}

	tcp = &TCPNetwork{
		listener:       nil,
		NetCM:          cm.NewCM(),
		UseHeaderLen:   true,
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

		RecvMode:  mode,
		SendRetry: 1,
	}

	err = tcp.listen(addr)
	if err != nil {
		tcp = nil
	}
	return
}

func NewEmptyTcp(mode int) *TCPNetwork {
	if HandlerProc <= 0 {
		HandlerProc = 0
	}

	return &TCPNetwork{
		listener:       nil,
		NetCM:          cm.NewCM(),
		UseHeaderLen:   true,
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

		RecvMode:  mode,
		SendRetry: 1,
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

	tcpAddr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return err
	}
	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	logger.INFO("grapeNet listen On:[%v]", bindAddr)

	c.listener = lis
	go c.onAccept()
	return nil
}

func (c *TCPNetwork) handleConn(conn *net.TCPConn) {
	// 设置TCP选项
	conn.SetKeepAlive(true)

	var client = NewConn(c, conn, c.CreateUserData())

	c.OnAccept(client)
	c.NetCM.Register <- client // 注册一个全局对象
	client.startProc()         // 启动线程
}

/// 连接池的处理
func (c *TCPNetwork) onAccept() {
	defer func() {
		if p := recover(); p != nil {
			stacks := utils.PanicTrace(4)
			panic := fmt.Sprintf("recover panics: %v call:%v", p, string(stacks))
			logger.ERROR(panic)
		}
	}()

	// 1000次错误 跳出去
	for failures := 0; failures < 1000; {
		tconn, listenErr := c.listener.AcceptTCP()
		if listenErr != nil {
			if ne, ok := listenErr.(net.Error); ok && ne.Temporary() {
				log.Printf("accept temp err: %v", ne)
				continue
			}

			logger.ERROR("accept error:%v", listenErr)
			failures++
			return
		}

		go c.handleConn(tconn)
	}
}

func (c *TCPNetwork) Runnable() {
	c.MainProc()
}
