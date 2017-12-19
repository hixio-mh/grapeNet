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
	hMap sync.Map
}

func NewMap() *FuncMap {
	return &FuncMap{}
}

func (m *FuncMap) Bind(cmd interface{}, fun MapHandler) error {
	if reflect.TypeOf(fun).Kind() != reflect.Func {
		return errors.New("handler must be a callable function")
	}

	m.hMap.Delete(cmd)
	m.hMap.Store(cmd, fun)
	return nil
}

func (m *FuncMap) buildCaller(cmd interface{}, args ...interface{}) (h interface{}, in []reflect.Value, err error) {
	err = nil
	h, ok := m.hMap.Load(cmd)
	if !ok {
		err = errors.New("Unknow Handler")
		return
	}

	t := reflect.TypeOf(h) // 获得对象类型
	argArr := []interface{}(args)

	// 数量和参数对不上
	if len(argArr) < t.NumIn() {
		err = errors.New("Not enough arguments")
		return
	}

	// 逐步压入参数
	in = make([]reflect.Value, t.NumIn()) //Panic if t is not kind of Func
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		if argType != reflect.TypeOf(argArr[i]) {
			err = errors.New(fmt.Sprintf("Value not found for type %v", argType))
			return
		}
		in[i] = reflect.ValueOf(argArr[i]) // 完成一个基本的CALL
	}

	return
}

// 带有返回参数
func (m *FuncMap) CallR(cmd interface{}, args ...interface{}) (result []interface{}, err error) {
	err = nil
	h, in, cerr := m.buildCaller(cmd, args...)
	if cerr != nil {
		err = cerr
		return
	}

	rv := reflect.ValueOf(h).Call(in)
	for _, rd := range rv {
		result = append(result, rd.Interface())
	}

	return
}

// 无返回参数
func (m *FuncMap) Call(cmd interface{}, args ...interface{}) error {
	h, in, err := m.buildCaller(cmd, args...)
	if err != nil {
		return err
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

func FastCallR(cmd interface{}, args ...interface{}) (result []interface{}, err error) {
	return defaultMap.CallR(cmd, args...)
}
