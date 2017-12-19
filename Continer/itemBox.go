// 常用于管理游戏背包、技能列表的容器
// 适用于手游、页游等多种类型
// 适用于格子类BOX容器
// 线程安全、适用于大并发场景
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/12/10
package continer

import (
	"sync"
)

const (
	EmptyGrid = "grid is empty..."
)

type ItemElement struct {
	// 位置信息
	col int
	row int

	// 格子是否为空
	isEmpty bool
	// 原始数据类型(例如数据信息)
	Value interface{}
	// 其他类型数据(例如描述信息)
	Other interface{}
}

func (e *ItemElement) Position() int {
	return (e.col * e.row)
}

func (e *ItemElement) Column() int {
	return e.col
}

func (e *ItemElement) Row() int {
	return e.row
}

func (e *ItemElement) IsEmpty() bool {
	return e.isEmpty
}

type ItemBox struct {
	// 标记属于谁,可以是任意类型
	Owner interface{}
	// 标记类型（建议自己定义）
	Type int

	// 有多少格子
	// 列
	column int
	// 行
	row int

	inter [][]ItemElement
	sync.RWMutex
}

////////////////////////////////////////////
// 创建函数
func NewBox(column, row, itype int, owner interface{}) *ItemBox {
	result := &ItemBox{
		Owner:  owner,
		Type:   itype,
		column: column,
		row:    row,
	}

	result.Init()

	return result
}

///////////////////////////////////////////
// 成员函数

func (b *ItemBox) Init() error {

}

// 用于保存
func (b *ItemBox) ToJson() (data string, err error) {

}

func (b *ItemBox) FromJson(src string) error {

}

func (b *ItemBox) ToBinary() (data []byte, err error) {

}

func (b *ItemBox) FromBinary(src []byte) error {

}

// 会产生数据COPY，如果需要组织数据包，请尽量避免使用该函数
func (b *ItemBox) Array() []ItemElement {

}

// 背包操作
// 道具数量
func (b *ItemBox) ItemCount() int {

}

// 格子的总数
func (b *ItemBox) GridCount() int {
	return (b.column * b.row)
}

// 空格子数量
func (b *ItemBox) EmptySpace() int {

}

// 是否已满
func (b *ItemBox) IsFull() bool {
	return (b.EmptySpace() == 0)
}

// 背包操作函数
// 自动压入到空格位置
func (b *ItemBox) Push(item, other interface{}) (err error, item *ItemElement) {

}

// 格子存在则放入其他格子并返回元素，没有空格子则返回失败
func (b *ItemBox) PushCell(col, row int, item, other interface{}) (err error, item *ItemElement) {

}

// 该格子存在道具那么把该格子道具放入一个空格子，如果没有没空格子，则返回失败
func (b *ItemBox) PushOrSwap(col, row int, item, other interface{}) error {

}

// 取出该格子道具
func (b *ItemBox) Peek(col, row int) *ItemElement {

}

// 取出该格子的VALUE数值，格子为空则返回错误
func (b *ItemBox) PeekValue(col, row int) (val interface{}, err error) {

}

// 取出该格子的Ohter数值，格子为空则返回错误
func (b *ItemBox) PeekOther(col, row int) (other interface{}, err error) {

}

func (b *ItemBox) Range(fn func(val *ItemElement) bool) {

}
