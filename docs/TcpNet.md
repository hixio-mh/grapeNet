## Tcp网络库基本用法

带有基本数据包打包通讯的基本TCPNet服务，轻量级TcpNet，支持粘包断包处理。

目前测试5000连接瞬发成功，应该可以支持更多，但是目前暂时没测试。

轻量级是为了更好的兼容`FlatBuffers`、`Msgpack`等序列化库。

示例参考: `samples/TcpNet`

### 监听服务

```
    tcpNet, err := tcp.NewTcpServer(":8799")
	if err != nil {
		log.Fatal(err)
		return
	}
```

非常简单，内建一个Connetion的管理器，支持broardcast行为，所以基本服务上比较容易，自动处理所有的粘黏包处理。

### 支持的Callback行为

```
    /// 所有的callBack函数

	// 创建用户DATA，创建用户关联数据，连接服务器后或用户连接进入后触发
	CreateUserData func() interface{}

	// 通知连接，当连接进入时触发
	OnAccept func(conn *TcpConn)
	// 数据包进入，数据包完成解密并触发处理
	OnHandler func(conn *TcpConn, ownerPak []byte)
	// 连接关闭，连接关闭时触发
	OnClose func(conn *TcpConn)
	// 连接成功，当连接服务器成功后触发，SERVER不会触发。
	OnConnected func(conn *TcpConn)

    // 简易主处理函数，用于处理自己服务器的逻辑【选用】
	MainProc func() 

	// 打包以及加密行为
    // 打包一个数据包，主要用于把任何类型转换为[]byte
	Package   func(val interface{}) (data []byte, err error)
    // 解包数据包，建议用默认
	Unpackage func(conn *TcpConn, spak *stream.BufferIO) (data [][]byte, err error)

	// 输出panic数据，当出现Panic或崩溃时触发
	Panic func(conn *TcpConn, src string)

    // 加密解密函数，打包的时候加密解密用的
	Encrypt func(data []byte) []byte
	Decrypt func(data []byte) []byte
```

所有Callback均提供默认的行为，所以可以选择自己需要的来处理。

### 连接服务器

```
    connNet := tcp.NewEmptyTcp() // 需要建立空的TcpNet

	connNet.OnHandler = RecvEchoMsg
	connNet.OnClose = OnClose
	// 连接建立
	_, err := connNet.Dial("127.0.0.1:8799", nil)
	if err != nil {
		log.Fatal(err)
	}
```

建立连接需要一个TcpNet，主要是因为需要自动使用其中的`连接管理器`，用于`broadcast`消息或快速查询`TcpConn`。

