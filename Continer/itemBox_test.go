package continer

import (
	"fmt"
	"testing"
)

type BoxItem struct {
	ItemId   int
	ItemName string
}

func Test_ItemBoxInit(t *testing.T) {
	itemBox, err := NewBox(10, 10, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}
	fmt.Println(*itemBox)

	itemBox, err = NewBox(10, 10, Inventory, 100, nil, nil)
	if err == nil {
		t.Fail()
		return
	}
}

func Test_ItemBoxSearch(t *testing.T) {
	itemBox, err := NewBox(100, 100, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	itemBox.Range(func(item *ItemElement) bool {
		fmt.Println(*item)
		return true
	})

	itemBox.Reverse(func(item *ItemElement) bool {
		fmt.Println(*item)
		return true
	})
}

func Test_ItemBoxFormatter(t *testing.T) {
	itemBox, err := NewBox(100, 100, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	itemBox.Push(&BoxItem{1000, "testitem"}, nil)
	itemBox.Push(&BoxItem{1002, "testitem"}, nil)
	itemBox.Push(&BoxItem{1003, "testitem"}, nil)
	itemBox.Push(&BoxItem{1004, "testitem"}, nil)

	out, err := itemBox.ToBinary()
	if err != nil {
		t.Fail()
		return
	}

	itemBox, err = NewBox(100, 100, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}
	err = itemBox.FromBinary(out)
	if err != nil {
		t.Fail()
		return
	}

}

func Test_ItemBoxPushAndPeek(t *testing.T) {
	itemBox, err := NewBox(100, 100, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	itemBox.PushCell(0, 0, &BoxItem{1000, "testitem"}, nil)
	itemBox.PushCell(1, 0, &BoxItem{1002, "testitem"}, nil)

	vitem, err := itemBox.Peek(0, 0)
	if err != nil {
		t.Fail()
		return
	}

	if vitem.x != 0 && vitem.y != 0 {
		t.Fail()
		return
	}

	val, err := itemBox.PeekValue(0, 0)
	if err != nil {
		t.Fail()
		return
	}

	if val != vitem.value {
		t.Fail()
		return
	}
}

func Test_ItemBoxMove(t *testing.T) {
	itemBox, err := NewBox(100, 100, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	itemBox.PushCell(0, 0, &BoxItem{1000, "testitem"}, nil)
	itemBox.PushCell(1, 0, &BoxItem{1002, "testitem"}, nil)
	itemBox.PushCell(1, 8, &BoxItem{1003, "testitem"}, nil)

	err = itemBox.Move(0, 0, 1, 2)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	sitem, _ := itemBox.Peek(0, 0)
	if sitem.IsEmpty() == false {
		t.Fail()
		return
	}

	sitem, _ = itemBox.Peek(1, 2)
	if sitem.IsEmpty() {
		t.Fail()
		return
	}

	err = itemBox.Move(1, 0, 1, 8)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	sitem, _ = itemBox.Peek(1, 8)
	val, _ := sitem.Value()
	if sitem.IsEmpty() || val.(*BoxItem).ItemId != 1002 {
		t.Fail()
		return
	}
}

func Test_ItemBoxRemove(t *testing.T) {
	itemBox, err := NewBox(100, 100, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	itemBox.PushCell(0, 0, &BoxItem{1000, "testitem"}, nil)
	itemBox.PushCell(1, 0, &BoxItem{1002, "testitem"}, nil)
	itemBox.PushCell(1, 8, &BoxItem{1003, "testitem"}, nil)

	err = itemBox.Remove(1, 0)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	sitem, _ := itemBox.Peek(1, 0)
	if sitem.IsEmpty() == false {
		t.Fail()
		return
	}

}

func Test_ItemBoxSort(t *testing.T) {
	itemBox, err := NewBox(10, 10, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	itemBox.PushCell(0, 0, &BoxItem{1000, "testitem333"}, nil)
	itemBox.PushCell(1, 0, &BoxItem{12003, "testitemasdasd"}, nil)
	itemBox.PushCell(1, 8, &BoxItem{1005, "testitemssss"}, nil)
	itemBox.PushCell(2, 3, &BoxItem{1001, "testitemddd"}, nil)
	itemBox.PushCell(4, 6, &BoxItem{1004, "testitem2222"}, nil)

	itemBox.Sort(func(av, bv interface{}) bool {
		return av.(*BoxItem).ItemId < bv.(*BoxItem).ItemId
	})

	arr := itemBox.Array()
	for _, v := range arr {
		vv, _ := v.Value()
		fmt.Println(v, vv)
	}
}

func Benchmark_PSPush(b *testing.B) {
	itemBox, err := NewBox(100, 100, Inventory, 100, &BoxItem{}, nil)
	if err != nil {
		b.Fail()
		fmt.Println(err)
		return
	}

	id := 1000
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			itemBox.Push(&BoxItem{id, "testitem333"}, nil)
			id++
		}
	})

	itemBox.Sort(func(av, bv interface{}) bool {
		return av.(*BoxItem).ItemId < bv.(*BoxItem).ItemId
	})

	arr := itemBox.Array()
	for _, v := range arr {
		vv, _ := v.Value()
		fmt.Println(v, vv)
	}
}
