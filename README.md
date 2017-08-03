## 简介（Introduction）

Go语言编写轻量级网络库 (grapeNet is a lightweight and Easy Use Network Framework)

可用于游戏服务端、强网络服务器端或其他类似应用场景，每个模块单独提取并且拥有独立的使用方法，内部耦合性较轻。

## 模块表（Function）

* Lua脚本绑定管理（可绑定任何类型的函数）
* 日志库（底层采用Seelog）
* 函数管理系统（可以根据任何类型参数将其与函数绑定并互相调用）
* 流处理
* Tcp网络
* Websocket网络* （未完成）
* Codec（任意类型注册对象并在其他位置动态创建该对象）

## 依赖第三方

* Seelog (github.com/cihub/seelog)
* Gopher-lua(github.com/yuin/gopher-lua)
* Gopher-luar(layeh.com/gopher-luar)

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

### 其他模块

暂无文档 有待补全