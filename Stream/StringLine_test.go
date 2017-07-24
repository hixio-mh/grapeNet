package grapeStream

import (
	"fmt"
	"testing"
)

func Test_Line(t *testing.T) {
	bSL := NewSL("MC 0 oG 2 0 t k gn3 ecV 8wn")
	t.Log(bSL.Source())
	fmt.Println(bSL, bSL.Command())

	bSL.Append("aasdasdasd")
	bSL.Append(1000)
	bSL.Append(111.2)

	bSL.AppendA62(12344)

	sPak := bSL.Pack()
	fmt.Println(sPak)
	t.Log(sPak)
}
