## Sgin快速生成库

用于在组合API请求时快速进行签名生成，自动根据动态参数组合Sign签名。

### 基本用法

```
// 声明一个测试对象 用于传递API参数等
type SignTest struct {
	TestAbc string `form:"abc" json:"jabc" sign:"abc"`
	DataAbc int    `form:"tint" json:"jtint" sign:"tint"`
	Time    int64  `form:"t" json:"jt" sign:"t"`
	Sign    string `form:"sign" b:"-" sign:"-"`
}
```

声明类型有各种类型，其中会用于sign传递的参数名为Sign，同时类型为string，那么我们需要对签名设置为"-"，意思为在打包整个签名中不计算该参数，现在我们来为这个结构生成签名。
> 注意：生成签名时不会计算sign的签名内容，且参数会自动排序。

签名排序规则为 根据字符首字母排序

我们来为结构生成MD5的签名

```
    SignTag = "sign" // 标记tag用于取结构后面的参数名
	SignKey = "20f0d253d40714277e5c12081db1237cafdc3999" // 签名用key

    // 赋值
	st := &SignTest{
		TestAbc: "123123asdasd",
		DataAbc: 30000,
		Time:    time.Now().Unix(),
	}

    // 生成一个MD5的签名
	sign, serr := SignMD5(st)
	if serr != nil {
		t.Error(serr)
		return
	}

	st.Sign = sign
```

除了MD5外，签名还支持SHA1，如果你需要更多算法，可以FORK我的项目再额外编辑添加。