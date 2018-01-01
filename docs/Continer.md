## syncList 带锁的List容器

一般用于需要多线程或多goroutine同步的情况下，带有读写所，底层为原始的list.List。
基本用于需要补充list无锁的场景下使用。

### 基本用法

```
    slist := continer.New()

    // 压入一个任意对象
    slist.Push(10000)
    // 弹出第一个
    val := slist.First()
    // 弹出最后一个
    val = slist.Back()
    // 删除全部
    slist.Clear()
    // 循环
    slist.Range(func(val interface{}) {
        // do something
    })
    // 查询
    slist.Search(func(val interface{}) bool {
        // do something
    })
    // 删除指定元素
    slist.Remove(func(val interface{}) bool {
        // do something
    })
```

## 游戏用背包容器 ItemBox

用于游戏内使用的背包，技能，穿戴装备等需要特殊容器的场景，自带线程安全！
由于涉及到游戏使用，所以相对于比较复杂。

> 分页场景较为复杂，项目不同，建议根据自己实际情况设计不同的分页代码。

> 全部代码已跑过test case以及benchmark

### 初始化背包

```
    // 假设背包大小 100 * 100的格子，类型为背包(Inventory),100为所属玩家ID，根据实际项目可以传入任何类型
    // BoxItem为游戏内实际的道具对象或技能对象，最后的参数为info，为附加参数，例如我司为ItemInfo结构，用于描述道具信息
	itemBox, err := NewBox(100, 100, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
```

### 压入一个道具

```
    // 道具自动放入背包的第一个空格，如果背包已满返回error
    itemBox.Push(&BoxItem{id, "testitem333"}, nil)
    // 道具放入指定位置 col = 1,row = 0
    itemBox.PushCell(1,0,&BoxItem{id, "testitem333"}, nil)
    // 道具放入指定位置并且如果该位置有道具，那么该道具自动放入一个空格，背包已满则返回错误
    itemBox.PushAndSwap(1,0,&BoxItem{id, "testitem333"}, nil)
```

### 取出道具

```
    // 取出指定格子的元素信息
    e,err := itemBox.Peek(1,0)
    // 取出指定格子道具的Value
    val,err := itemBox.PeekValue(1,0)
    // 取出指定格子道具的Info
    info,err := itemBox.PeekInfo(1,0)
```

### 排序函数

```
    // 用于对背包进行排序,返回true时交换2个格子位置
    itemBox.Sort(func(av, bv interface{}) bool {
		return av.(*BoxItem).ItemId < bv.(*BoxItem).ItemId
	})
```

### 导出和导入为[]byte

```
    // 导出背包数据为[]byte
    // 注意只导出非空格子数据，且使用bson进行数据打包，value以及info
	out, err := itemBox.ToBinary()
	if err != nil {
		return
	}

    // 导入数据
    err = itemBox.FromBinary(out)
	if err != nil {
		return
	}
```

### 其他函数

```
    // 删除道具
    itemBox.Remove(1,0)
```

更多函数请参考源码或test case源文件。