## ETCD简易封装模块（针对Watch的）

一个针对etcd v3客户端的简易封装，封装的主要目的是为了简化Put和Get操作中的复杂内容，增加序列化和反序列的更多快速选项，减少代码。
同时增加针对Watch行为的多Key监听，对于需要处理更多复杂参数有效隔离开，并针对Type等做处理（自动推导任意参数，但首参数必须为固定类型）。

> 由于有些同学有特殊需求，针对Watcher增加Key的Callback必须绑定参数，这下完美了。

#### 连接

``` go
	import (
		etcd "github.com/koangel/Etcd"
	)


	err := etcd.Dial([]string{"localhost:2379"})
	if err != nil {
		// do something
		return
	}
```

#### 读写任意对象

可以通过函数 `SetFormatter` 来设定自己的格式化方法，四个函数都需要必须实现，内建格式json和bson
> bson格式化自动转换为base64，所以必须实现`ToString`以及`FromString`，否则解析可能存在问题。

> 默认不设置自动使用Json格式。

```go
 etcd.SetFormatter(&JsonFormatter{})
```

压入对象以及取出对象
```go
	err := etcd.MarshalKey("fooObj", etcd.M{
		"abcd":  "strings",
		"int":   3000,
		"float": 1.234,
	})

	if err != nil {
		fmt.Print(err)
		return
	}

	var uMap etcd.M = etcd.M{}
	err = etcd.UnmarshalKey("fooObj", &uMap)
	if err != nil {
		fmt.Print(err)
		return
	}
```

#### 租约与服务发现

可以快速序列化一个租约键值，可用于服务发现或配置记录等行为，当然我基本用于服务发现。

> 服务发现的一个简单例子

```Go
// 压入一个服务，并开启一个持续续约的系统，一旦过期主服务器认为服务已关闭
 Id,err := etcd.MarshalKeyTTL(
	 "game_server",etcd.M{
		 "server":"127.0.0.1",
	 },
	 60)

 if err != nil {
	fmt.Print(err)
	return
 }

 // 持续启动续约
 etcd.Keeplive(Id)

 // 强制过期
 etcd.Revoke(Id)

```

#### 启用一个Watch监控

callback必须声明

第一个参数 `vtype string` 用于接收Watch类型

第二个参数 `key []byte` 用于接收Watch时的Key,由于特殊情况下依然需要Key，例如监控某个目录的话，子项目的改变需要KEY来确定。

第三个参数 `values []byte` 用于接收Watch时的Value

```go

func TestHandler(vtype string,key,values []byte,other01 int,other02 string,other03 float32) {
	// do something
}

```

快速建立一个监听

```go
err := etcd.BindWatcher("/fooWatch",TestHandler,1000,"other02_test",1.0)
if err != nil {
	fmt.Print(err)
	return
}
```

每当对`fooWatch`做出任何操作，底层自动call TestHandler。
