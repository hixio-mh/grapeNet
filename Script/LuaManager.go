// 基于github.com/yuin/gopher-lua的lua管理器
// 实现基本的热更新理论 lua层
// 可做配置 可做逻辑脚本
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/9
package grapeLua

import (
	"fmt"
	"sync"
)

type LuaManager struct {
	luaMap map[string]*LuaVM

	locker sync.Mutex
}

var Ins LuaManager = LuaManager{
	luaMap: make(map[string]*LuaVM),
}

///////////////////////////////////
// 创建一个lua脚本
func NewFromFile(name, filename string) *LuaVM {
	newLua := &LuaVM{
		l:           nil,
		ScriptName:  name,
		LuaFileName: filename,
		LuaData:     "",
	}

	newLua.New(name)
	err := newLua.DoFile(filename)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	Ins.addVM(name, newLua)
	return newLua
}

// 来自DATA的不支持热更新
func NewFromData(name, luaData string) *LuaVM {
	newLua := &LuaVM{
		l:           nil,
		ScriptName:  name,
		LuaFileName: "",
		LuaData:     luaData,
	}

	newLua.New(name)
	err := newLua.DoString(luaData)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	Ins.addVM(name, newLua)
	return newLua
}

func NewVM(name string) *LuaVM {
	newLua := &LuaVM{
		l:           nil,
		ScriptName:  name,
		LuaFileName: "",
		LuaData:     "",
	}

	newLua.New(name)
	Ins.addVM(name, newLua)
	return newLua
}

///////////////////////////////////
// 管理工具
func (m *LuaManager) addVM(name string, vm *LuaVM) {
	m.locker.Lock()
	defer m.locker.Unlock()

	_, ok := m.luaMap[name]
	if ok {
		delete(m.luaMap, name)
	}

	m.luaMap[name] = vm
}

func (m *LuaManager) Find(name string) *LuaVM {
	m.locker.Lock()
	defer m.locker.Unlock()

	vm, ok := m.luaMap[name]
	if ok {
		return vm
	}

	return nil
}

func (m *LuaManager) Call(name string, fnName string, args ...interface{}) {
	lvm := m.Find(name)
	if lvm == nil {
		return
	}

	lvm.CallGlobal(fnName, args)
}

func (m *LuaManager) BindToAll(fnName string, v interface{}) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, v := range m.luaMap {
		v.SetGlobal(fnName, v)
	}
}
