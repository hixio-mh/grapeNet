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
	//fmt.Println(s, a, f)
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

func Benchmark_Call(b *testing.B) {
	lua := NewFromFile("testlua", "../_lua_tests/luascripts/call_lua_test.lua")
	if lua == nil {
		b.Error("unknow lua")
		return
	}

	for i := 0; i < b.N; i++ {
		err := lua.CallGlobal("TestAbc", "a", 2)
		if err != nil {
			b.Error(err)
			return
		}

		// call ret
		_, rerr := lua.CallGlobalRet("TestAbcRet", 20, 33)
		if rerr != nil {
			b.Error(rerr)
			return
		}
	}
}

func Benchmark_CallParallel(b *testing.B) {
	lua := NewFromFile("testlua", "../_lua_tests/luascripts/call_lua_test.lua")
	if lua == nil {
		b.Error("unknow lua")
		return
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := lua.CallGlobal("TestAbc", "a", 2)
			if err != nil {
				b.Error(err)
				return
			}

			// call ret
			_, rerr := lua.CallGlobalRet("TestAbcRet", 20, 33)
			if rerr != nil {
				b.Error(rerr)
				return
			}
		}
	})
}
