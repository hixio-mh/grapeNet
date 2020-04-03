 [![Go Report Card](https://goreportcard.com/badge/github.com/koangel/grapeNet)](https://goreportcard.com/report/github.com/koangel/grapeNet)  [![Build Status](https://secure.travis-ci.org/koangel/grapeNet.png)](http://travis-ci.org/koangel/grapeNet)

---

## 简介（Introduction）

Go语言编写轻量级网络库 (grapeNet is a lightweight and Easy Use Network Framework)

可用于游戏服务端、强网络服务器端或其他类似应用场景，每个模块单独提取并且拥有独立的使用方法，内部耦合性较轻。

其实GO语言曾经有过很多强架构的框架，比如GoWorld之类的，已经足够了，但是我会将库用于各种轻量级应用不需要过于复杂的内容，所以我设计了GrapeNet，目的是模块独立化。
你可以拆开只使用其中很小的模块，也可以组合成一个服务端，并且在架构中设计也较为轻松，至于热更新的问题，目前脚本数据支持热更新，并且是自动的，只要跑一下UPDATE即可，程序本身稍后测试后发布（仅支持LINUX）。

本库更像是一个日常服务端开发的轻量级工具库集合，用的开心噢。

本库内的大部分子模块均用于实际线上游戏产品、防御类产品以及支付类产品中，经过一定的检验，可以放心使用。

慢慢更新中，很多坑要填，根据近期测试，除LUA库和网络库外，其他库均可直接用于商业产品。

个人博客：http://grapec.me/

> 注意：由于ETCD V3 CLIENT不支持1.9以下版本的GO环境，所以ETCD库不在对1.8以下版本支持，TravisCI的BUILD状态也不支持1.9以下版本。
> 仅仅支持 Go 1.10以及以上版本。

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
* Kcp网络(基于KCPGO的网络基础库)
* Websocket网络 （基础版，兼容版）
* Websocketv2网络实验版 （使用zero-copy upgrader，更低的损耗和更好效率，不兼容v1版，开启压缩）
* Codec（任意类型注册对象并在其他位置动态创建该对象）
* CSV序列化模块（通过Tag可以直接序列化到对象或对象序列化为CSV）
* Sign生成库（自动将结构或map[string]interface{}排序后生成一个sign，可以自行设置KEY）
* Etcd简易封装，针对Watcher做任意参数的监听callback(多Key监听)
* Continers容器库，游戏用背包容器、带有锁的并行LIST等
* Utils多种简易辅助库的集合（三元运算符、数值转换、轻并行执行库、启动Daemon）

## 依赖第三方

* Seelog (github.com/cihub/seelog)
* Gopher-lua(github.com/yuin/gopher-lua)
* Gopher-luar(layeh.com/gopher-luar)
* Websocket旧版使用 (github.com/gorilla/websocket)
* Websocketv2使用 (github.com/gobwas/ws)
* Etcd ClientV3(github.com/coreos/etcd)
* Bson (gopkg.in/mgo.v2/bson)
* Daemon (github.com/takama/daemon)
* Kcp-Go (github.com/xtaci/kcp-go)

不依赖任何CGO内容，lua本身也是纯GO实现。

## 文档（Documentation）

### 基本用法文档

* [CSV Doc](./docs/CSV.md)
* [FuncMap Doc](./docs/FuncMap.md)
* [Codec Doc](./docs/Codec.md)
* [Lua Doc](./docs/LuaScript.md)
* [Etcd Doc](./docs/Etcd.md)
* [Sign Doc](./docs/Sign.md)
* [Utils Doc](./docs/Utils.md)
* [Continer Doc](./docs/Continer.md)
* [KcpNet Doc](./docs/KcpNet.md)
* [TcpNet Doc](./docs/TcpNet.md)
* [WSNet Doc](./docs/WSNet.md)
* [WSNetV2（实验版） Doc](./docs/WSNetV2.md)

其他文档我会陆陆续续逐步完善并补充。

### 其他模块

暂无文档 有待补全