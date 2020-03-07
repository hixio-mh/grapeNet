## Kcp网络库基本用法

带有基本数据包打包通讯的基本KcpNet服务，轻量级KcpNet，支持粘包断包处理。

基于最基本的KCP库开发而成，主要是尝试KCP在游戏领域以及其他服务器端开发领域的应用。

做了5000的基本测试，内存使用略微有点高，稍后近一步在项目中测试。

轻量级是为了更好的兼容`FlatBuffers`、`Msgpack`等序列化库。

##### 2020-03-07

新增独立的数据包处理线程 与 IO分离 防止因处理导致的IO超时，同时多个协程进行处理。

设置处理数据量为：
```
kcpNet.HandlerProc = 2
```

默认数量为2

示例参考: `samples/KcpNet`

### 监听服务

```
    cnf := kcpNet.NewConfig()
  	cnf.Mode = "aes"
  
  	// 需要自主加解密算法的使用
  	//cnf.Mode = "none"
  
  	kcpNetwork,err := kcpNet.NewKcpServer(":4744",cnf)
  	if err != nil {
  		log.Fatal(err)
  	}
  
  	kcpNetwork.OnAccept = func(conn *kcpNet.KcpConn) {
  		log.Printf("Kcp Accept In:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr().String())
  	}
  	
  	kcpNetwork.OnHandler = func(conn *kcpNet.KcpConn, ownerPak []byte) {
  		log.Printf("Kcp Accept In:%v From:%v Recv:%v", conn.SessionId, conn.TConn.RemoteAddr().String(),string(ownerPak))
  	}
  	
  	kcpNetwork.OnClose = func(conn *kcpNet.KcpConn) {
  		if conn.TConn == nil {
  			return
  		}
  
  		log.Printf("Default Closed:%v From:%v", conn.SessionId, conn.TConn.RemoteAddr)
  	}
  
  	kcpNetwork.Runnable()
```

非常简单，内建一个Connetion的管理器，支持broardcast行为，所以基本服务上比较容易，自动处理所有的粘黏包处理。

### 支持的Callback行为

```
    /// 所有的callBack函数

	// 创建用户DATA，创建用户关联数据，连接服务器后或用户连接进入后触发
	CreateUserData func() interface{}

	// 通知连接，当连接进入时触发
	OnAccept func(conn *KcpConn)
	// 数据包进入，数据包完成解密并触发处理
	OnHandler func(conn *KcpConn, ownerPak []byte)
	// 连接关闭，连接关闭时触发
	OnClose func(conn *KcpConn)
	// 连接成功，当连接服务器成功后触发，SERVER不会触发。
	OnConnected func(conn *KcpConn)

    // 简易主处理函数，用于处理自己服务器的逻辑【选用】
	MainProc func() 

	// 打包以及加密行为
    // 打包一个数据包，主要用于把任何类型转换为[]byte
	Package   func(val interface{}) (data []byte, err error)
    // 解包数据包，建议用默认
	Unpackage func(conn *KcpConn, spak *stream.BufferIO) (data [][]byte, err error)

	// 输出panic数据，当出现Panic或崩溃时触发
	Panic func(conn *KcpConn, src string)

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

建立连接需要一个KcpNet，主要是因为需要自动使用其中的`连接管理器`，用于`broadcast`消息或快速查询`TcpConn`。

### 发送数据包

发送数据包一共有2种体系，一种是广播例如：

```
    // 广播字节码
    kcpNetwork.NetCM.Broadcast([]byte(fmt.Sprintf("this is echo msg:%v", i)))
```

还有一种就是单独针对CONN发送数据，例如：

```
    // 发送字节码，不简易但提供该函数
    conn.Send([]byte("test"))

    // 发送一个对象，会调用Callback Package函数
    // 目前会转换成BSON对象的[]byte，可以自定义如何转换。
    conn.SendPak(object{1,2,"test"})
```

### 接收数据

只需要绑定KcpNet中的Callback，OnHandler函数即可，会传入收到数据的Conn以及全部包内容，自行处理即可。