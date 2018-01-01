package formatters

import (
	"fmt"
	"testing"
)

type FMTTest struct {
	Abcd string
	DDD  float32
	CCCC int
}

func Test_BsonFormatter(t *testing.T) {
	var fte ItemFormatter = new(BsonFormatter)

	out, err := fte.To(&FMTTest{
		Abcd: "test string...",
		DDD:  0.333333,
		CCCC: 100000,
	}, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	tempv := new(FMTTest)
	err = fte.From(out, tempv, nil)
	if err != nil {
		t.Fail()
		fmt.Println(err)
		return
	}

	fmt.Println(*tempv)
}
