 [![Go Report Card](https://goreportcard.com/badge/github.com/koangel/grapeNet)](https://goreportcard.com/report/github.com/koangel/grapeNet)  [![Build Status](https://secure.travis-ci.org/koangel/grapeNet.png)](http://travis-ci.org/koangel/grapeNet)

---

## 简介（Introduction）

Go语言编写轻量级网络库 (grapeNet is a lightweight and Easy Use Network Framework)

可用于游戏服务端、强网络服务器端或其他类似应用场景，每个模块单独提取并且拥有独立的使用方法，内部耦合性较轻。

其实GO语言曾经有过很多强架构的框架，比如GOWOLRD之类的，已经足够了，但是我会将库用于各种轻量级应用不需要过于复杂的内容，所以我设计了GrapeNet，目的是模块独立化。
你可以拆开只使用其中很小的模块，也可以组合成一个服务端，并且在架构中设计也较为轻松，至于热更新的问题，目前脚本数据支持热更新，并且是自动的，只要跑一下UPDATE即可，程序本身稍后测试后发布（仅支持LINUX）。

本库更像是一个日常服务端开发的轻量级工具库集合，用的开心噢。

慢慢更新中，很多坑要填，根据近期测试，除LUA库和网络库外，其他库均可直接用于商业产品。

个人博客：http://grapec.me/

## 安装

```go
go get -u github.com/koangel/grapeNet/...
```

## 模块表（Function）

* Lua脚本绑定管理（可绑定任何类型的函数、线程安全且自动推倒类型）
* 日志库（底层采用Seelog）
* 函数管理系统（可以根据任何类型参数将其与函数绑定并互相调用）
* 流处理
* Tcp网络
* Websocket网络 （基础版）
* Codec（任意类型注册对象并在其他位置动态创建该对象）
* CSV序列化模块（通过Tag可以直接序列化到对象或对象序列化为CSV）
* Sign生成库（自动将结构或map[string]interface{}排序后生成一个sign，可以自行设置KEY）
* Etcd简易封装，针对Watcher做任意参数的监听callback(多Key监听)

## 依赖第三方

* Seelog (github.com/cihub/seelog)
* Gopher-lua(github.com/yuin/gopher-lua)
* Gopher-luar(layeh.com/gopher-luar)
* Etch ClientV3(github.com/coreos/etcd)
* Bson (gopkg.in/mgo.v2/bson)

不依赖任何CGO内容，lua本身也是纯GO实现。

## 文档（Documentation）

### Lua脚本库使用

Lua库为线程安全库，可以在任意协程中并行调用脚本文件中的函数，也可以合并脚本库。

#### 新建文件脚本

```go
    lua := NewFromFile("testlua", "../_lua_tests/luascripts/call_lua_test.lua")
```

#### 新建字符串脚本（用于加密LUA文件或内存脚本）

```go
    lua := NewVM("testRegister")
    lua.DoString(sScript) // 执行字符串
```

#### 获取脚本中的结构

```go
	person := Person{}
	err := lua.GetTable("person", &person)
	if err != nil {
		t.Error(err)
		return
	}
```

#### 绑定与GO互相调用

脚本模块允许Go函数与Lua脚本无缝调用且线程安全。

> 注意：部分代码请参考LuaCall_test.go中的代码

```go
    // 绑定本地GO函数
    lua := NewVM("testRegister")
	lua.SetGlobal("TestGoFunc", bindTestFn)

    // 一定要在DO文件之前绑定，否则调用该文件时可能无效
    // 调用Lua中函数

    err := lua.CallGlobal("TestAbc", "a", 2)
	if err != nil {
		return
	}
```

### Codec模块介绍

该模块用于动态注册一个类型并于某个字符串或类型绑定，稍后使用相同的字符串构建该类型。

#### 基本用法
```go
	R("VTest", VTObject{}) // 注册对象VTObject

	// 通过字符串构建对象
	obj, err := New("VTest")
```
> 注意：详细代码可参考Codec_test.go

### FuncMap模块（Function Map）

该模块用于动态绑定某个函数至MAP中并通过任意类型传参并调用该函数。
（可绑定任意参数数量以及参数类型的函数，且可以与任何类型的绑定，线程安全且高效率）
（可以在自己的机器上跑跑我的benchmark）

#### 绑定函数

```go
	// 假定需要绑定函数为(任意参数)
	func vestAbc(s string, i int)...

	func vest3(i uint32, s string, data string)...
```

```go

	// 与字符串绑定
	FastBind("0", vestAbc)

	// 与Int类型绑定
	FastBind(1, vestAbc)

	// 再次绑定一个不同参数数量以及参数类型的函数
	FastBind(2.0, vest3)
```

#### 调用绑定的函数

```go
	// 调用与字符串0关联函数，第一参数为 CALL 0字符串，第二个参数为 1233
	FastCall("0", "Call 0", 1233)

	// 调用以int关联的函数
	FastCall(1, "Call 1", 2000)
	
```

> 注意：具体代码请参考FuncMap下的FuncMap_test.go内容，含benchmark

### CSV序列化库

可以将某一行直接序列化为结构或结构序列化为数据（线程安全），50W写出仅需要2S，读入50W数据也仅仅需要2S

#### 新建CSV类
```go
	// 构建一个默认的CSV，Token为 ',' 且自动忽略第一行数据为Header
	newCSV, err := NewCSVDefault("../_csv_tests/test001.csv")
	if err != nil ...
```

```go
	// 构建一个自定义的CSV Token为Tab,且自动忽略第一行数据为Header
	newCSV, err := NewCSV("../_csv_tests/test001.csv", '	', true)
	if err != nil ...
```

#### 将某一行序列化为STRUCT

```go
	type testCsvNode struct {
		Name     string  `column:"name"`
		LastName string  `column:"tags"`
		Data     float32 `column:"data"`
		Value    int     `column:"dataval"`
	}

	// 直接读取0行数据到对象中 
	sval := &testCsvNode{}
	newCSV.GetRow(0, sval)
```

#### 添加一行数据
```go

	// Append添加一个对象数据数据
	for i := 0; i < 2000; i++ {
		newCSV.Append(testCsvNode{
			Name:     fmt.Sprintf("name:%d", i),
			LastName: fmt.Sprintf("temp:%d", i+1000),
			Data:     1.2222 + float32(i),
			Value:    i * 2000,
		})
	}

	newCSV.SaveAll() // 保存
```
#### 设置头数据
```go

	// 新建文件
	newCSV, err := CreateCSV("../_csv_tests/Benchmark001.csv", Default_token, testCsvNode{})
	if err != nil {
		b.Error(err)
		return
	}
```
```go
	// 根据结构自动设置头
	newCSV.SetHeader(testCsvNode{})
```

### ETCD简易封装模块（针对Watch的）

一个针对etcd v3客户端的简易封装，封装的主要目的是为了简化Put和Get操作中的复杂内容，增加序列化和反序列的更多快速选项，减少代码。
同时增加针对Watch行为的多Key监听，对于需要处理更多复杂参数有效隔离开，并针对Type等做处理（自动推导任意参数，但首参数必须为固定类型）。


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
	err := MarshalKey("fooObj", map[string]interface{}{
		"abcd":  "strings",
		"int":   3000,
		"float": 1.234,
	})

	if err != nil {
		fmt.Print(err)
		return
	}

	var uMap map[string]interface{} = map[string]interface{}{}
	err = UnmarshalKey("fooObj", &uMap)
	if err != nil {
		fmt.Print(err)
		return
	}
```

#### 启用一个Watch监控

callback必须声明
第一个参数 vtype string 用于接收Watch类型
第二个参数 values []byte 用于接收Watch时的Value，由于KEY在声明时已最齐，这里就先不传入了。

```go

func TestHandler(vtype string,values []byte,other01 int,other02 string,other03 float32) {
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


### 其他模块

暂无文档 有待补全