package kcpNet

import (
	"github.com/xtaci/kcp-go"

	"fmt"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
	stream "github.com/koangel/grapeNet/Stream"
)

type KcpNetwork struct {
	listener *kcp.Listener

	NetCM *cm.ConnManager

	KcpConf *KcpConfig

	RecvMode int

	KcpBlock kcp.BlockCrypt
	/// 所有的callBack函数
	// 创建用户DATA
	CreateUserData func() interface{}

	// 通知连接
	OnAccept func(conn *KcpConn)
	// 数据包进入
	OnHandler func(conn *KcpConn, ownerPak []byte)
	// 连接关闭
	OnClose func(conn *KcpConn)
	// 连接成功
	OnConnected func(conn *KcpConn)

	MainProc func() // 简易主处理函数

	// 打包以及加密行为
	Package   func(val interface{}) (data []byte, err error)
	Unpackage func(conn *KcpConn, spak *stream.BufferIO) (data [][]byte, err error)

	// 输出panic数据
	Panic func(conn *KcpConn, src string)

	Encrypt func(data, key []byte) []byte
	Decrypt func(data, key []byte) []byte

	// ping,pong CALL
	SendPing func(conn *KcpConn)
	SendPong func(conn *KcpConn, ping []byte) bool
}

var (
	HandlerProc = 1
)

/////////////////////////////
// 创建网络服务器
func parserConf(config *KcpConfig) (*KcpConfig, kcp.BlockCrypt) {
	switch config.Mode {
	case "normal":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 0, 40, 2, 1
	case "fast":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 0, 30, 2, 1
	case "fast2":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 1, 20, 2, 1
	case "fast3":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 1, 10, 2, 1
	}

	var block kcp.BlockCrypt
	pass := []byte(config.CryptKey)
	switch config.Crypt {
	case "sm4":
		block, _ = kcp.NewSM4BlockCrypt(pass[:16])
	case "tea":
		block, _ = kcp.NewTEABlockCrypt(pass[:16])
	case "xor":
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	case "none":
		block, _ = kcp.NewNoneBlockCrypt(pass)
	case "aes-128":
		block, _ = kcp.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		block, _ = kcp.NewAESBlockCrypt(pass[:24])
	case "blowfish":
		block, _ = kcp.NewBlowfishBlockCrypt(pass)
	case "twofish":
		block, _ = kcp.NewTwofishBlockCrypt(pass)
	case "cast5":
		block, _ = kcp.NewCast5BlockCrypt(pass[:16])
	case "3des":
		block, _ = kcp.NewTripleDESBlockCrypt(pass[:24])
	case "xtea":
		block, _ = kcp.NewXTEABlockCrypt(pass[:16])
	case "salsa20":
		block, _ = kcp.NewSalsa20BlockCrypt(pass)
	default:
		config.Crypt = "aes"
		block, _ = kcp.NewAESBlockCrypt(pass)
	}

	return config, block
}

func NewKcpServer(addr string, cnf *KcpConfig) (tcp *KcpNetwork, err error) {

	config, block := parserConf(cnf)

	if HandlerProc <= 0 {
		HandlerProc = 0
	}

	tcp = &KcpNetwork{
		listener: nil,
		NetCM:    cm.NewCM(),

		KcpConf:  config,
		KcpBlock: block,

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

		RecvMode: cnf.Recvmode,
	}

	err = tcp.listen(addr)
	if err != nil {
		logger.ERRORV(err)
		tcp = nil
	}
	return
}

func NewEmptyKcp(cnf *KcpConfig) *KcpNetwork {
	config, block := parserConf(cnf)

	if HandlerProc <= 0 {
		HandlerProc = 0
	}

	return &KcpNetwork{
		listener: nil,
		NetCM:    cm.NewCM(),

		KcpConf:  config,
		KcpBlock: block,

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
		RecvMode: cnf.Recvmode,
	}
}

////////////////////////////
// 成员函数
func (c *KcpNetwork) RemoveSession(sessionId string) {
	c.NetCM.Remove(sessionId)
}

func (c *KcpNetwork) Dial(addr string, UserData interface{}) (conn *KcpConn, err error) {
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

func (c *KcpNetwork) listen(bindAddr string) error {
	if c.listener != nil {
		return fmt.Errorf("listener is nil...")
	}

	config := c.KcpConf

	lis, err := kcp.ListenWithOptions(bindAddr, c.KcpBlock, config.DataShard, config.ParityShard)
	if err != nil {
		return err
	}

	if err := lis.SetDSCP(config.DSCP); err != nil {
		return err
	}

	if err := lis.SetReadBuffer(config.SockBuf); err != nil {
		return err
	}

	if err := lis.SetWriteBuffer(config.SockBuf); err != nil {
		return err
	}

	logger.INFO("kcpNetwork listen On:[%v]", bindAddr)

	c.listener = lis
	go c.onAccept()
	return nil
}

/// 连接池的处理
func (c *KcpNetwork) onAccept() {
	defer func() {
		if p := recover(); p != nil {
			logger.ERROR("recover panics: %v", p)
		}
	}()
	// 1000次错误 跳出去
	for failures := 0; failures < 1000; {
		conn, listenErr := c.listener.AcceptKCP()
		if listenErr != nil {
			logger.ERROR("accept error:%v", listenErr)
			failures++
			continue
		}

		failures = 0
		logger.INFO("New Kcp Connection:%v，Accept.", conn.RemoteAddr())
		var client = NewConn(c, conn, c.CreateUserData())

		c.OnAccept(client)

		c.NetCM.Register <- client // 注册一个全局对象

		client.startProc() // 启动线程
	}
}

func (c *KcpNetwork) Runnable() {
	c.MainProc()
}
