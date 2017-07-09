// 基于github.com/yuin/gopher-lua的lua管理器
// 实现基本的热更新理论 lua层
// 可做配置 可做逻辑脚本
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/9
package grapeLua

import (
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
)

///////////////////////////////////////
// 绑定相关参数
// 可以绑定函数，类等
// 本来打算自己写发现有库
func (vm *LuaVM) SetGlobal(fnName string, fn interface{}) {
	if vm.l == nil {
		return
	}

	vm.l.SetGlobal(fnName, luar.New(vm.l, fn))
}

///////////////////////////////////////
// 弱类型的调用指定的脚本函数
func (vm *LuaVM) CallGlobal(fnName string, args ...interface{}) (err error) {
	if vm.l == nil {
		err = errors.New("lua state error...")
		return
	}

	fn := vm.l.GetGlobal(fnName)
	if fn.Type() != lua.LTFunction {
		err = errors.New(fmt.Sprintf("Unknow Lua Function:%v", fnName))
		return
	}

	// 组合参数列表
	lpValues := []lua.LValue{}
	argsArr := []interface{}(args)
	for _, v := range argsArr {
		lpValues = append(lpValues, luar.New(vm.l, v))
	}

	err = vm.l.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}, lpValues...)
	return
}

func (vm *LuaVM) CallGlobalRet(fnName string, args ...interface{}) (r lua.LValue, err error) {
	if vm.l == nil {
		err = errors.New("lua state error...")
		return
	}

	fn := vm.l.GetGlobal(fnName)
	if fn.Type() != lua.LTFunction {
		err = errors.New(fmt.Sprintf("Unknow Lua Function:%v", fnName))
		return
	}

	// 组合参数列表
	lpValues := []lua.LValue{}
	argsArr := []interface{}(args)
	for _, v := range argsArr {
		lpValues = append(lpValues, luar.New(vm.l, v))
	}

	err = vm.l.CallByParam(lua.P{
		Fn:      fn,
		NRet:    3,
		Protect: true,
	}, lpValues...)

	r = vm.l.Get(-1)
	vm.l.Pop(1)
	return
}
