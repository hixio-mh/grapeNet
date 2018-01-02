package Utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"

	"go.uber.org/zap/buffer"
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

//快速压缩一个数据内存 不支持大数据，不要用来压缩较大文件
func FastGZipMsg(src []byte, isBase64 bool) (data []byte, err error) {
	var zipBuf buffer.Buffer
	zap := gzip.NewWriter(&zipBuf)

	_, err = zap.Write(src)
	if err != nil {
		return
	}

	err = zap.Close()
	if err != nil {
		return
	}

	data = zipBuf.Bytes()
	if isBase64 {
		data = []byte(base64.StdEncoding.EncodeToString(data))
	}

	return
}

// 快速解压一个消息 不支持大数据，不要用来解压缩较大文件
func FastUnGZipMsg(src []byte, isBase64 bool) (data []byte, err error) {
	bzip := src
	if isBase64 {
		decode, berr := base64.StdEncoding.DecodeString(string(bzip))
		if berr != nil {
			err = berr
			return
		}

		bzip = decode
	}

	zr, gerr := gzip.NewReader(bytes.NewReader(bzip))
	if gerr != nil {
		err = gerr
		return
	}

	unzipByte, uerr := ioutil.ReadAll(zr)
	if uerr != nil {
		err = uerr
		return
	}

	zr.Close()

	data = unzipByte
	return
}
