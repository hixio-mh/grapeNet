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

	"github.com/yuin/gluamapper"
	"github.com/yuin/gopher-lua"
)

// 将TABLE数据反射到一个STRUCT中
// 得到一个全局的结构
func (vm *LuaVM) GetTable(gName string, val interface{}) error {
	vm.mux.Lock()
	defer vm.mux.Unlock()

	rTable := vm.l.GetGlobal(gName)
	if rTable.Type() != lua.LTTable {
		return errors.New("Data Type Error...")
	}

	return gluamapper.Map(rTable.(*lua.LTable), val)
}
