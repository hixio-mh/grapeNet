// 自动泛型任意参数CALL
// 参考部分inject代码
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/25

package grapeFunc

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type MapHandler interface{}

type FuncMap struct {
	hMap map[interface{}]MapHandler

	locker sync.Mutex
}

func NewMap() *FuncMap {
	return &FuncMap{
		hMap: make(map[interface{}]MapHandler),
	}
}

func (m *FuncMap) Bind(cmd interface{}, fun MapHandler) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	if reflect.TypeOf(fun).Kind() != reflect.Func {
		return errors.New("handler must be a callable function")
	}

	_, ok := m.hMap[cmd]
	if ok {
		delete(m.hMap, cmd) // 删除旧的
	}

	m.hMap[cmd] = fun

	return nil
}

func (m *FuncMap) Call(cmd interface{}, args ...interface{}) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	h, ok := m.hMap[cmd]
	if !ok {
		return errors.New("Unknow Handler")
	}

	t := reflect.TypeOf(h) // 获得对象类型
	argArr := []interface{}(args)

	// 数量和参数对不上
	if len(argArr) < t.NumIn() {
		return errors.New("Not enough arguments")
	}

	// 逐步压入参数
	var in = make([]reflect.Value, t.NumIn()) //Panic if t is not kind of Func
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		if argType != reflect.TypeOf(argArr[i]) {
			return errors.New(fmt.Sprintf("Value not found for type %v", argType))
		}
		in[i] = reflect.ValueOf(argArr[i]) // 完成一个基本的CALL
	}

	reflect.ValueOf(h).Call(in)
	return nil
}

// 默认行为
var defaultMap = NewMap()

// 快速绑定
func FastBind(cmd interface{}, fun MapHandler) error {
	return defaultMap.Bind(cmd, fun)
}

// 快速调用
func FastCall(cmd interface{}, args ...interface{}) error {
	return defaultMap.Call(cmd, args...)
}
