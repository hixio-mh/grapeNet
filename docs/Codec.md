## Codec模块介绍

该模块用于动态注册一个类型并于某个字符串或类型绑定，稍后使用相同的字符串构建该类型。

#### 基本用法
```go
	R("VTest", VTObject{}) // 注册对象VTObject

	// 通过字符串构建对象
	obj, err := New("VTest")
```
> 注意：详细代码可参考Codec_test.go