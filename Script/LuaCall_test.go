package grapeLua

import (
	"fmt"
	"testing"
)

func Test_callTest(t *testing.T) {
	lua := NewFromFile("testlua", "../_lua_tests/luascripts/call_lua_test.lua")
	if lua == nil {
		t.Error("lua do file error...")
		return
	}
	err := lua.CallGlobal("TestAbc", "a", 2)
	if err != nil {
		t.Error(err)
		return
	}

	// call ret
	ret, rerr := lua.CallGlobalRet("TestAbcRet", 20, 33)
	if rerr != nil {
		t.Error(rerr)
		return
	}

	fmt.Println(ret)

	t.Log("call Finished...")
}

func bindTestFn(s string, a int, f float32) {
	fmt.Println(s, a, f)
}
func Test_callRegister(t *testing.T) {
	lua := NewVM("testRegister")
	lua.SetGlobal("TestGoFunc", bindTestFn)

	err := lua.DoFile("../_lua_tests/luascripts/call_go_test.lua") // 运行
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("call Finished...")
}
