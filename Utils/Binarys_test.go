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

func TestGzip(t *testing.T) {
	mergeBuf := MergeBinary([]byte("i am first words"), []byte("i am second words"))

	gzip, err := FastGZipMsg(mergeBuf, true)
	if err != nil {
		t.Fail()
		return
	}

	unzip, err := FastUnGZipMsg(gzip, true)
	if err != nil {
		t.Fail()
		return
	}

	splitBuf := SplitBinary(unzip)

	for _, v := range splitBuf {
		fmt.Println(string(v))
	}
}
