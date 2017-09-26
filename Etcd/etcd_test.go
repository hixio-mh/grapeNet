package grapeEtcd

import (
	"fmt"
	"testing"
)

func Test_etcd_dial(t *testing.T) {
	err := Dial([]string{"localhost:2379"})
	if err != nil {
		t.Error(err)
		return
	}
}

func Test_etcd_Object(t *testing.T) {
	err := Dial([]string{"localhost:2379"})
	if err != nil {
		t.Error(err)
		return
	}

	SetFormatter(&JsonFormatter{})

	err = MarshalKey("fooObj", map[string]interface{}{
		"abcd":  "strings",
		"int":   3000,
		"float": 1.234,
	})

	if err != nil {
		t.Error(err)
		return
	}

	var uMap map[string]interface{} = map[string]interface{}{}
	err = UnmarshalKey("fooObj", &uMap)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Print(uMap)
}
