// 自动泛型任意参数CALL
// 中间会PANIC，依赖MARTINI的INJECT
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/25

package grapeFunc

import (
	"errors"
	"reflect"
	"sync"

	"github.com/koangel/grapeNet/inject"
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

func (m *FuncMap) Register(cmd interface{}, fun MapHandler) error {
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

	inval := inject.New()
	argArr := []interface{}(args)
	for _, v := range argArr {
		inval.Map(v)
	}

	inval.Invoke(h)
	return nil
}
