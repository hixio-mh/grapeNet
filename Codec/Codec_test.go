package grapeCodec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type VTObject struct {
	Name string
	VAO  int
	Data string
}

func Test_Codec(t *testing.T) {

	TempV := &VTObject{
		Name: "asdasd",
		VAO:  2000,
		Data: "azxzxczxc",
	}
	v, verr := json.Marshal(TempV)
	if verr != nil {
		t.Error(verr)
		return
	}

	RA(VTObject{})
	R("VTest", VTObject{})
	obj, err := New("VTest")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(obj, reflect.TypeOf(obj))
	if err = json.Unmarshal(v, obj); err != nil {
		t.Error(err)
		return
	}

	fmt.Println(obj)
}

func BenchmarkCodec(b *testing.B) {
	TempV := &VTObject{
		Name: "asdasd",
		VAO:  2000,
		Data: "azxzxczxc",
	}
	v, verr := json.Marshal(TempV)
	if verr != nil {
		return
	}

	RA(VTObject{})
	R("VTest", VTObject{})

	for i := 0; i < b.N; i++ {
		obj, err := New("VTest")
		if err != nil {
			return
		}

		//fmt.Println(obj, reflect.TypeOf(obj))
		if err = json.Unmarshal(v, obj); err != nil {
			return
		}

		//fmt.Println(obj)
	}
}
