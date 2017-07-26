// 自动根据注册类型根据协议结构反射对应类型
// 线程安全类型
// 通过类型名称快速New一个对象指针
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/26
package grapeCodec

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var objmap map[string]reflect.Type = make(map[string]reflect.Type)
var locker sync.Mutex

// register auto
func RA(val interface{}) {
	t := reflect.TypeOf(val)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	R(fmt.Sprintln(t), val)
}

// register
func R(name string, val interface{}) {
	locker.Lock()
	defer locker.Unlock()

	_, ok := objmap[name]
	if ok {
		delete(objmap, name)
	}

	objmap[name] = reflect.TypeOf(val)
}

func New(name string) (o interface{}, err error) {
	locker.Lock()
	defer locker.Unlock()
	err = nil
	o = nil
	if v, ok := objmap[name]; ok {
		o = reflect.New(v).Interface()
	} else {
		err = errors.New("Unregister Codec Object...")
	}

	return
}
