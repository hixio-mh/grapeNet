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
	"fmt"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"

	"gopkg.in/mgo.v2/bson"

	formatter "github.com/koangel/grapeNet/Continer/formatters"
	util "github.com/koangel/grapeNet/Utils"
)

const (
	EmptyGrid = "grid is empty..."
)

const (
	Inventory = iota
	Wear
	Skill
)

type ItemElement struct {
	// 位置信息
	x   int
	y   int
	pos int // 为了效率 直接存一个更直接

	// 格子是否为空
	isEmpty bool
	// 原始数据类型(例如数据信息)
	value interface{}
	// 其他类型数据(例如描述信息)
	info interface{}

	locker sync.RWMutex
}

func (e *ItemElement) Position() int {
	return e.pos
}

func (e *ItemElement) Column() int {
	return e.x
}

func (e *ItemElement) Row() int {
	return e.y
}

func (e *ItemElement) IsEmpty() bool {
	return e.isEmpty
}

func (e *ItemElement) SetValue(value, info interface{}) {
	e.locker.Lock()
	defer e.locker.Unlock()

	e.isEmpty = false
	e.value = value
	e.info = info
}

func (e *ItemElement) Remove() {
	e.locker.Lock()
	defer e.locker.Unlock()

	e.isEmpty = true
	e.value = nil
	e.info = nil
}

func (e *ItemElement) Value() (value, info interface{}) {
	e.locker.RLock()
	defer e.locker.RUnlock()

	value = e.value
	info = e.info
	return
}

type ItemBox struct {
	// 标记属于谁,可以是任意类型
	Owner interface{}
	// 标记类型（建议自己定义）,例如 Inventory,wear,skill
	Type int
	// 有多少格子
	// 列
	maxcolumn int
	// 行
	maxrow int
	// count
	count int32
	empty int32

	vtype reflect.Type
	vinfo reflect.Type

	inter []*ItemElement
	once  sync.Once
	l     sync.RWMutex

	formatter formatter.ItemFormatter
}

////////////////////////////////////////////
// 创建函数
func NewBox(column, row, itype int, owner, value, info interface{}) (box *ItemBox, err error) {
	box = nil
	err = nil
	if value == nil {
		err = fmt.Errorf("value is nill.")
		return
	}

	if owner == nil {
		err = fmt.Errorf("owner is nill.")
		return
	}

	box = &ItemBox{
		Owner:     owner,
		Type:      itype,
		maxcolumn: column,
		maxrow:    row,
		vtype:     nil,
		vinfo:     nil,
		formatter: new(formatter.BsonFormatter),
	}

	box.registerType(value, info)
	box.Init()
	box.calc()

	return
}

///////////////////////////////////////////
// 成员函数
func (b *ItemBox) Init() {
	b.once.Do(func() {
		b.inter = make([]*ItemElement, b.maxcolumn*b.maxrow)
		x, y := 0, 0

		for i := range b.inter {
			if y >= b.maxrow {
				break
			}

			if x >= b.maxcolumn {
				y++
				x = 0
			}

			b.inter[i] = &ItemElement{
				isEmpty: true,
				value:   nil,
				info:    nil,
				// x = col ,y = row
				x: x,
				y: y,
				// pos  = (row * maxrow) + col
				pos: b.pos(x, y),
			}

			x++
		}
	})
}

func (b *ItemBox) SetFormatter(foramtter formatter.ItemFormatter) {
	b.formatter = foramtter
}

// 用于保存
// 保存协议,|count|len|Element Body|len|Element Body|
func (b *ItemBox) ToBinary() (data []byte, err error) {
	data = []byte{}

	array := [][]byte{}
	b.l.RLock()
	defer b.l.RUnlock()

	for _, item := range b.inter {
		out, ferr := b.formatter.To(item.value, item.info)
		if ferr != nil {
			continue
		}
		ElementData := bson.M{
			"x":     item.x,
			"y":     item.y,
			"empty": item.isEmpty,
			"body":  out,
		}

		bt, berr := bson.Marshal(ElementData)
		if berr != nil {
			err = berr
			return
		}

		array = append(array, bt)
	}

	data = util.MergeBinary(array...)
	err = nil
	return
}

func (b *ItemBox) FromBinary(src []byte) error {
	out := util.SplitBinary(src)

	b.l.Lock()
	defer b.l.Unlock()

	for _, src := range out {
		ed := bson.M{}
		err := bson.Unmarshal(src, &ed)
		if err != nil {
			return err
		}

		x := util.MustInt(ed["x"], -1)
		y := util.MustInt(ed["y"], -1)
		pos := b.pos(x, y)

		if x < 0 || y < 0 {
			return fmt.Errorf("Position is Error...")
		}

		if pos >= len(b.inter) {
			return fmt.Errorf("Position is Error...")
		}

		item := b.inter[pos]
		item.x = x
		item.y = y
		item.pos = pos
		item.isEmpty = util.MustBool(ed["empty"], false)

		data, ok := ed["body"]
		if !ok {
			return fmt.Errorf("It's Unknow Format...")
		}

		value := reflect.New(b.vtype).Interface()
		var info interface{}
		if b.vinfo != nil {
			info = reflect.New(b.vinfo).Interface()
		}

		err = b.formatter.From(data.([]byte), value, info)
		if err != nil {
			return err
		}

		item.value = value
		item.info = info
	}

	b.calc()

	return nil
}

// 会产生数据COPY，如果需要组织数据包，请尽量避免使用该函数
// 仅仅打包已存在数据，空格不会打包
func (b *ItemBox) Array() []ItemElement {
	result := []ItemElement{}
	b.l.RLock()
	defer b.l.RUnlock()
	for _, item := range b.inter {
		if item.IsEmpty() == false {
			result = append(result, *item)
		}
	}
	return result
}

// 背包操作
// 道具数量
func (b *ItemBox) ItemCount() int {
	return int(b.count)
}

// 格子的总数
func (b *ItemBox) GridCount() int {
	return (b.maxcolumn * b.maxrow)
}

// 空格子数量
func (b *ItemBox) EmptyGrid() int {
	return int(b.empty)
}

// 是否已满
func (b *ItemBox) IsFull() bool {
	return (b.EmptyGrid() == 0)
}

// 获得一个空格子位置信息
func (b *ItemBox) EmptyGird() (col, row int) {
	col, row = -1, -1
	b.l.RLock()
	defer b.l.RUnlock()

	for _, item := range b.inter {
		if item.IsEmpty() {
			col, row = item.x, item.y
			return
		}
	}
	return
}

// 背包操作函数
// 暂不提供批量压入
// 自动压入到空格位置
func (b *ItemBox) Push(item, info interface{}) (err error, ri *ItemElement) {
	ri = nil
	err = nil

	col, row := b.EmptyGird()
	if col == -1 || row == -1 {
		err = fmt.Errorf("box is full...")
		return
	}

	err, ri = b.PushCell(col, row, item, info)
	return
}

// 格子存在则放入其他格子并返回元素，没有空格子则返回失败
func (b *ItemBox) PushCell(col, row int, item, info interface{}) (err error, ri *ItemElement) {
	ri = nil
	err = nil

	if b.IsFull() {
		err = fmt.Errorf("box is full...")
		return
	}

	if reflect.TypeOf(item).Name() != b.vtype.Name() {
		err = fmt.Errorf("value type is error...")
		return
	}

	pos := b.pos(col, row)
	if pos >= len(b.inter) {
		err = fmt.Errorf("position out of range...")
		return
	}

	b.l.RLock()
	vi := b.inter[pos]
	b.l.RUnlock()

	if vi.IsEmpty() == false {
		err = fmt.Errorf("grid is not empty...")
		return
	}

	ri = vi
	ri.SetValue(item, info)

	// 重新计算格子数据
	b.calc()

	return
}

// 该格子存在道具那么把该格子道具放入一个空格子，如果没有没空格子，则返回失败
func (b *ItemBox) PushAndSwap(col, row int, item, info interface{}) error {
	if reflect.TypeOf(item).Name() != b.vtype.Name() {
		return fmt.Errorf("value type is error...")
	}

	if b.IsFull() {
		return fmt.Errorf("box is full...")
	}

	pos := b.pos(col, row)
	if pos >= len(b.inter) {
		return fmt.Errorf("position out of range...")
	}

	b.l.RLock()
	vi := b.inter[pos]
	b.l.RUnlock()

	if !vi.isEmpty {
		ex, ey := b.EmptyGird()

		// 将数据放入一个空格子
		b.Move(vi.x, vi.y, ex, ey) // 移动格子

	}

	vi.SetValue(item, info)
	return nil
}

// 取出该格子道具
func (b *ItemBox) Peek(col, row int) (item *ItemElement, err error) {
	item = nil
	err = nil
	pos := b.pos(col, row)
	if pos >= len(b.inter) || col < 0 || row < 0 {
		err = fmt.Errorf("position out of range...")
		return
	}

	b.l.RLock()
	defer b.l.RUnlock()
	item = b.inter[pos]
	return
}

// 取出该格子的VALUE数值，格子为空则返回错误
func (b *ItemBox) PeekValue(col, row int) (val interface{}, err error) {
	item, perr := b.Peek(col, row)
	if perr != nil {
		err = perr
		return
	}

	if item.isEmpty {
		err = fmt.Errorf("grid is empty.")
		return
	}

	val = item.value

	return
}

// 取出该格子的Ohter数值，格子为空则返回错误
func (b *ItemBox) PeekInfo(col, row int) (info interface{}, err error) {
	item, perr := b.Peek(col, row)
	if perr != nil {
		err = perr
		return
	}

	if item.isEmpty {
		err = fmt.Errorf("grid is empty.")
		return
	}

	info = item.info
	return
}

// 移动道具 A to b,如果b or a都不为空那么交换
func (b *ItemBox) Move(sc, sr, dc, dr int) error {

	spos := b.pos(sc, sr)
	dpos := b.pos(dc, dr)
	if spos >= len(b.inter) || dpos >= len(b.inter) {
		return fmt.Errorf("position out of range...")
	}

	b.l.RLock()
	defer b.l.RUnlock()

	sitem := b.inter[spos]
	ditem := b.inter[dpos]
	if sitem.isEmpty && ditem.isEmpty {
		return fmt.Errorf("both of grids is empty...")
	}

	if sitem.isEmpty {
		return fmt.Errorf("source grid is empty...")
	}

	val, info := sitem.value, sitem.info // 先把数据取出来
	if ditem.isEmpty {
		sitem.Remove()
		ditem.SetValue(val, info)
	} else {
		sitem.SetValue(ditem.value, ditem.info)
		ditem.SetValue(val, info)
	}

	return nil
}

func (b *ItemBox) SwapElement(se, de *ItemElement) error {
	return b.Move(se.x, se.y, de.x, de.y)
}

func (b *ItemBox) Remove(col, row int) error {
	pos := b.pos(col, row)
	if pos >= len(b.inter) || col < 0 || row < 0 {
		return fmt.Errorf("position out of range...")
	}

	b.l.RLock()
	defer b.l.RUnlock()
	item := b.inter[pos]
	if item.isEmpty {
		return fmt.Errorf("grid is empty...")
	}

	item.Remove()

	return nil
}

// 函数返回true那么交换2个格子位置
// 排序会把前面空格自动填满
func (b *ItemBox) Sort(fn func(av, bv interface{}) bool) {
	if b.IsFull() == false {
		c, r := b.EmptyGird()
		pos := b.pos(c, r)

		for i, item := range b.inter {
			if item.isEmpty == false && i > pos {
				b.Move(item.x, item.y, c, r)

				c, r = b.EmptyGird()
				if c == -1 || r == -1 {
					break
				}
				pos = b.pos(c, r)
			}
		}
	}

	sortSlice := b.Array()
	// 开始排序
	sort.Slice(sortSlice, func(i, j int) bool {
		return fn(sortSlice[i].value, sortSlice[j].value)
	})

	b.l.Lock()
	for i, si := range sortSlice {
		b.inter[i].SetValue(si.value, si.info)
	}
	b.l.Unlock()
}

// 正向迭代
func (b *ItemBox) Range(fn func(val *ItemElement) bool) {
	b.l.RLock()
	defer b.l.RUnlock()

	for _, item := range b.inter {
		if fn(item) == false {
			break
		}
	}
}

// 反向迭代
func (b *ItemBox) Reverse(fn func(val *ItemElement) bool) {
	b.l.RLock()
	defer b.l.RUnlock()

	for i := len(b.inter) - 1; i > 0; i-- {
		if fn(b.inter[i]) == false {
			break
		}
	}
}

/////////////////////////////
// inter
func (b *ItemBox) pos(col, row int) int {
	return (row * b.maxrow) + col
}

func (b *ItemBox) registerType(val, info interface{}) {
	if val != nil && b.vtype == nil {
		b.vtype = reflect.TypeOf(val)
	}

	if info != nil && b.vinfo == nil {
		b.vinfo = reflect.TypeOf(info)
	}
}

func (b *ItemBox) calc() {
	count, empty := 0, 0
	for _, item := range b.inter {
		if item.IsEmpty() {
			empty++
		} else {
			count++
		}
	}

	atomic.StoreInt32(&b.count, int32(count))
	atomic.StoreInt32(&b.empty, int32(empty))
}
