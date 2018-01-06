package grapeStream

import (
	"fmt"
	"testing"
)

var allPacket = []byte("testasdasdasdddddddddddddddddddddddddddddddddddddddasdasd")

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

	for i := 0; i < 1000; i++ {
		unPackSm.WriteAuto(writeBytes) // 写进去
	}

	_, uerr := unPackSm.Unpack(defaultFn)
	if uerr != nil {
		t.Error(uerr)
		return
	}

	t.Log("finished...")
}

func Benchmark_Unpack(b *testing.B) {
	unPackSm := NewPacker()
	for i := 0; i < b.N; i++ {
		writeBytes, err := PackerOnce([]byte(fmt.Sprintf("asdasddddddddd %v", i)), defaultFn)
		if err != nil {
			b.Error(err)
			return
		}

		unPackSm.WriteAuto(writeBytes) // 写进去
		if i%2 == 0 {
			pak, uerr := unPackSm.Unpack(defaultFn)
			if uerr != nil {
				break
			}

			if pak == nil {
				break
			}
			unPackSm.Reset()
		}
	}
}

func defaultDecrypt(data []byte) []byte {
	return data
}

func Test_UnpackLine(t *testing.T) {
	var stream BufferIO

	stream.WriteAuto([]byte("asdasddddddddddddddddddddd\nasdasddddddddddddddddddddd\n"))

	pakData := [][]byte{}

	for {
		pData, err := stream.UnpackLine(defaultDecrypt)
		if err != nil {
			break
		}

		if pData == nil {
			break
		}

		pakData = append(pakData, pData)
	}
}
