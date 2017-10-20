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

var objmap sync.Map

// register auto
func RA(val interface{}) {
	t := reflect.TypeOf(val)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	R(fmt.Sprint(t), val)
}

// register
func R(name string, val interface{}) {

	// 目前对load or store的定义不详，暂时采用先Delete再store的形式
	objmap.Delete(name)
	objmap.Store(name, reflect.TypeOf(val))
}

func New(name string) (o interface{}, err error) {
	err = nil
	o = nil
	if v, ok := objmap.Load(name); ok {
		o = reflect.New(v.(reflect.Type)).Interface()
	} else {
		err = errors.New("Unregister Codec Object...")
	}

	return
}
