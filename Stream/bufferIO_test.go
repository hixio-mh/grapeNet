package grapeStream

import (
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

	_, err := pack.Packer(defaultFn)
	if err != nil {
		t.Error(err)
		return
	}
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
	_, uerr := unPackSm.Unpack(true, defaultFn)
	if uerr != nil {
		t.Error(uerr)
		return
	}

	t.Log("finished...")
}
func defaultDecrypt(data []byte) []byte {
	return data
}

func Test_UnpackLine(t *testing.T) {
	var stream BufferIO

	stream.WriteAuto([]byte("asdasddddddddddddddddddddd\nasdasddddddddddddddddddddd\n"))

	pakData := [][]byte{}

	for {
		pData, err := stream.UnpackLine(true, defaultDecrypt)
		if err != nil {
			break
		}

		if pData == nil {
			break
		}

		pakData = append(pakData, pData)
	}
}
