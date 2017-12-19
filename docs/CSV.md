## CSV序列化库

可以将某一行直接序列化为结构或结构序列化为数据（线程安全），50W写出仅需要2S，读入50W数据也仅仅需要2S

### 新建CSV类
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

### 将某一行序列化为STRUCT

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

### 添加一行数据
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
### 设置头数据
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