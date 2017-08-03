package grapeLua

import (
	"fmt"
	"testing"
)

type Role struct {
	Name string
}

type Person struct {
	Name      string
	Age       int
	WorkPlace string
	Role      []*Role
}

func Test_GetLuaData(t *testing.T) {
	lua := NewFromFile("getData", "../_lua_tests/luascripts/get_data_test.lua")
	if lua == nil {
		t.Error("lua do file error...")
		return
	}

	person := Person{}
	err := lua.GetTable("person", &person)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(person)
	t.Log("finished...")
}

func Benchmark_LuaData(b *testing.B) {
	lua := NewFromFile("getData", "../_lua_tests/luascripts/get_data_test.lua")
	if lua == nil {
		b.Error("lua do file error...")
		return
	}

	for i := 0; i < b.N; i++ {
		person := Person{}
		err := lua.GetTable("person", &person)
		if err != nil {
			b.Error(err)
			return
		}
	}
}

func Benchmark_LuaData_Parallel(b *testing.B) {
	lua := NewFromFile("getData", "../_lua_tests/luascripts/get_data_test.lua")
	if lua == nil {
		b.Error("lua do file error...")
		return
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			person := Person{}
			err := lua.GetTable("person", &person)
			if err != nil {
				b.Error(err)
				return
			}
		}
	})
}
