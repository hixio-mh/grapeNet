## FuncMap模块（Function Map）


该模块用于动态绑定某个函数至MAP中并通过任意类型传参并调用该函数。
（可绑定任意参数数量以及参数类型的函数，且可以与任何类型的绑定，线程安全且高效率）

> 可以在自己的机器上跑跑我的benchmark

### What's News

* 2017/12/18

    (+)add: 增加带有返回参数的快速调用

### 绑定函数

```go
	// 假定需要绑定函数为(任意参数)
	func vestAbc(s string, i int)...

	func vest3(i uint32, s string, data string)...

	func vest4Result(i uint32, s string, data string) (uint32, string) {
		return i, s
	}
```

```go

	// 与字符串绑定
	FastBind("0", vestAbc)

	// 与Int类型绑定
	FastBind(1, vestAbc)

	// 再次绑定一个不同参数数量以及参数类型的函数
	FastBind(2.0, vest3)
```

### 调用绑定的函数

```go
	// 调用与字符串0关联函数，第一参数为 CALL 0字符串，第二个参数为 1233
	FastCall("0", "Call 0", 1233)

	// 调用以int关联的函数
	FastCall(1, "Call 1", 2000)
	
	// 调用带有返回参数的函数，并将返回参数以数组的形式返回
	res, err := FastCallR("0Result", uint32(3000), "Call_Float", "zxxczxcxc")
	if err != nil {
		fmt.Println(err)
		return
	}
```

> 注意：具体代码请参考FuncMap下的FuncMap_test.go内容，含benchmark