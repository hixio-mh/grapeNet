package Utils

import (
	"fmt"
	"testing"
)

func TestBinary(t *testing.T) {
	mergeBuf := MergeBinary([]byte("i am first words"), []byte("i am second words"))

	splitBuf := SplitBinary(mergeBuf)

	for _, v := range splitBuf {
		fmt.Println(string(v))
	}
}
