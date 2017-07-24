package grapeStream

import (
	"fmt"
	"testing"
)

var allPacket = []byte("testasdasdasdasdasd")

func defaultFn(val []byte) []byte {
	return val
}

func Test_Resize(t *testing.T) {
	pack := BuildPacker(allPacket)
	if pack == nil {
		t.Error("build pack error...")
		return
	}

	for i := 0; i < 1000; i++ {
		pack.WriteInt32(int32(i))
		pack.WriteAuto(allPacket)
	}

	newPack, err := pack.Packer(defaultFn)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(newPack)
}

func Test_Unpack(t *testing.T) {
	pack := BuildPacker(allPacket)
	if pack == nil {
		t.Error("build pack error...")
		return
	}

	writeBytes, err := pack.Packer(defaultFn)
	if err != nil {
		t.Error(err)
		return
	}

	unPackSm := NewPacker()
	unPackSm.WriteAuto(writeBytes) // 写进去
	newUnpack, uerr := unPackSm.Unpack(true, defaultFn)
	if uerr != nil {
		t.Error(uerr)
		return
	}

	fmt.Println(newUnpack)

	t.Log("finished...")
}
