package Utils

import "testing"

func Test_Ternary(t *testing.T) {
	checkV := false
	checkV2 := true

	str := Ifs(checkV, "isTrue", "isFalse")
	if str != "isFalse" {
		t.Fail()
		return
	}

	str = Ifs(checkV, "isTrue", Ifs(checkV2, "chk2Ture", "chk2False"))
	if str != "chk2Ture" {
		t.Fail()
		return
	}
}
