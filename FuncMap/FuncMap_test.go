package grapeFunc

import (
	"fmt"
	"testing"
)

func vestAbc(s string, i int) {
	fmt.Println(s, i)
}

func vest3(i uint32, s string, data string) {
	fmt.Println(i, s, data)
}

func Test_MapCall(t *testing.T) {
	FunMap := NewMap()
	FunMap.Bind("0", vestAbc)
	FunMap.Bind(1, vestAbc)

	FunMap.Bind(2.0, vest3)
	FunMap.Bind("CCCC", vest3)

	FunMap.Call("0", "Call 0", 1233)
	FunMap.Call("0", "Call 0 Sc", 1233, "asdasd", 4444)
	FunMap.Call(1, "Call 1", 2000)

	FunMap.Call("CCCC", uint32(2000), "asdasd", "zxxczxcxc")
	FunMap.Call(2.0, uint32(3000), "Call_Float", "zxxczxcxc")

}
