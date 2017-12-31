package Utils

import (
	"bytes"
	"encoding/binary"
)

/// 合并多个[]byte为一个[]bytes
/// 协议 |count|len|bytes body|len|bytes body|

/// Merge Multiple []byte arrary to []byte
/// protocol : |count|len|bytes body|len|bytes body|

func MergeBinary(src ...[]byte) []byte {
	array := [][]byte(src)
	buffer := new(bytes.Buffer)
	count := int32(len(array))

	err := binary.Write(buffer, binary.BigEndian, count) // 写入数量
	if err != nil {
		return []byte{}
	}
	for _, buf := range array {
		bufflen := int32(len(buf))
		binary.Write(buffer, binary.BigEndian, bufflen)
		buffer.Write(buf)
	}

	return buffer.Bytes()
}

/// 拆分 合并后的[]bytes
func SplitBinary(src []byte) [][]byte {
	buffer := bytes.NewBuffer(src)
	count := int32(0)

	binary.Read(buffer, binary.BigEndian, &count)

	result := [][]byte{}
	for i := 0; i < int(count); i++ {
		blen := int32(0)
		binary.Read(buffer, binary.BigEndian, &blen)
		buf := make([]byte, blen)
		_, err := buffer.Read(buf)
		if err != nil {
			return result
		}
		result = append(result, buf)
	}

	return result
}
