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
	"os"
	"time"

	"github.com/yuin/gopher-lua"
)

type LuaVM struct {
	l *lua.LState // Lua

	ScriptName  string // 脚本名称
	LuaFileName string // 文件名
	LuaData     string // 脚本数据

	lastModTime time.Time
}

//////////////////////////////////
// lua的执行函数
func (vm *LuaVM) New(name string) {
	if vm.l == nil {
		vm.l = lua.NewState()
		vm.ScriptName = name
	}
}

// 刷新脚本内容
// 自动根据文件名以及文件的修改日期刷LUA文件
func (vm *LuaVM) Update() {
	if len(vm.LuaFileName) > 1 {
		fi, err := os.Stat(vm.LuaFileName)
		if err != nil {
			return
		}

		if vm.lastModTime.Equal(fi.ModTime()) {
			return
		}

		vm.Close() // 关闭当前的数据
		vm.New(vm.ScriptName)
		vm.DoFile(vm.LuaFileName)
	}
}

func (vm *LuaVM) DoString(s string) error {
	if vm.l != nil {
		return vm.l.DoString(s)
	}

	return errors.New("lua state error...")
}

func (vm *LuaVM) DoFile(s string) error {
	if vm.l != nil {
		vm.LuaFileName = s
		fi, err := os.Stat(s)
		if err == nil {
			vm.lastModTime = fi.ModTime()
		}
		return vm.l.DoFile(s)
	}

	return errors.New("lua state error...")
}

func (vm *LuaVM) Close() {
	if vm.l != nil {
		vm.l.Close()
		vm.l = nil
	}
}

func (vm *LuaVM) State() *lua.LState {
	return vm.l
}

//////////////////////////////////////
// 对返回类型的封装
