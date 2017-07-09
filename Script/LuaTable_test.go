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
