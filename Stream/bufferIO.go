// 用于网络数据传输的高性能STREAM类
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/8

package grapeStream

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	defaultSize = 2048 // 默认数据包长度
)

type BufferIO struct {
	vBuffer []byte //缓存数据

	readIndex  int64 // 读入长度
	writeIndex int64 // 写入长度
}

type CryptFn func(val, key []byte) []byte

////////////////////////////////////////////////////////
// 新建一个Stream
func NewPacker() *BufferIO {
	return &BufferIO{
		vBuffer:    make([]byte, defaultSize),
		readIndex:  0,
		writeIndex: 0, // 预留4个字节的包长度
	}
}

func NewResize(size int) *BufferIO {
	return &BufferIO{
		vBuffer:    make([]byte, size),
		readIndex:  0,
		writeIndex: 0, // 预留4个字节的包长度
	}
}

func BuildResize(body []byte) (buf *BufferIO) {
	buf = NewResize(len(body) + 8)
	buf.WriteInt32(int32(len(body))) // 写入包长度
	buf.WriteAuto(body)
	return
}

func BuildPacker(body []byte) (buf *BufferIO) {
	buf = NewPacker()
	buf.WriteInt32(int32(len(body))) // 写入包长度
	buf.WriteAuto(body)
	return
}

///////////////////////////////////////////////////////
// 成员函数
// 得到有效的长度 缓冲区长度
func (b *BufferIO) Len() int64 {
	return int64(len(b.vBuffer))
}

// 读到结尾了
func (b *BufferIO) EndOf() bool {
	return (b.Len() - b.readIndex) <= 0
}

// 可读取的剩余长度
func (b *BufferIO) Available() int64 {
	if b.writeIndex > int64(len(b.vBuffer)) {
		return 0
	}

	return (b.writeIndex - b.readIndex)
}

// 返回缓冲区全部数据
func (b *BufferIO) Bytes() []byte {
	return b.vBuffer
}

// 扩容行为
func (b *BufferIO) Resize(newSize int64) error {
	if newSize < b.Len()+defaultSize {
		return errors.New("NewSize is too short...")
	}

	iResize := int64(cap(b.vBuffer)*2 + defaultSize)
	if newSize > defaultSize*2 {
		iResize = newSize + int64(cap(b.vBuffer))
	}

	tmpData := make([]byte, iResize) // 保存构建一个新的BUFFER
	copy(tmpData, b.vBuffer)         // copy旧的数据进
	b.vBuffer = tmpData              // 新的缓冲区大小

	return nil
}

func (b *BufferIO) Slice(startIndex, endIndex int) []byte {
	if int64(startIndex) >= b.Len() || startIndex == endIndex {
		return b.vBuffer
	}

	if int64(endIndex) > b.Len() {
		return b.vBuffer[startIndex:]
	}

	return b.vBuffer[startIndex:endIndex]
}

// 跳过指定的字节数量
// 数据自动前移
func (b *BufferIO) Shift(size int) error {
	if int64(size) > b.writeIndex {
		return errors.New("Shift Size is too big...")
	}
	copy(b.vBuffer, b.vBuffer[size:]) // 后面的数据返回到前面
	b.readIndex -= int64(size)
	b.writeIndex -= int64(size)

	return nil
}

func (b *BufferIO) Reset() {
	b.writeIndex = 0
	b.readIndex = 0
	b.vBuffer = make([]byte, defaultSize*2)
}

// 相当于不去读取而是改变pos位置
func (b *BufferIO) Skip(size int) error {
	if int64(size) > b.writeIndex {
		return errors.New("Skip Size is too big...")
	}

	b.readIndex += int64(size) // 跳过部分字节
	return nil
}

// 从开始跳到指定位置
func (b *BufferIO) Seek(pos int64) error {
	if int64(pos) > b.writeIndex {
		return errors.New("Seek Pos OverFlow...")
	}
	if pos < 0 {
		return errors.New("Unknow Seek Pos...")
	}

	b.readIndex = int64(pos)
	return nil
}

func (b *BufferIO) WSeek(pos int64) error {
	if int64(pos) > int64(len(b.vBuffer)) {
		return errors.New("Seek Pos OverFlow...")
	}
	if pos < 0 {
		return errors.New("Unknow Seek Pos...")
	}

	b.writeIndex = int64(pos)
	return nil
}

// 获得当前的读位置
func (b *BufferIO) ReadPos() int64 {
	return b.readIndex
}

// 获取当前写入位置
func (b *BufferIO) WritePos() int64 {
	return b.writeIndex
}

// 读行为
// 取出数据 但不计数
func (b *BufferIO) PeekBytes(size int) []byte {
	if int64(size) > b.Available() {
		return []byte{}
	}
	if int64(size) > b.Available() {
		return []byte{}
	}

	if int(b.readIndex+int64(size)) > len(b.vBuffer) {
		fmt.Println(int(b.readIndex+int64(size)), int64(size), len(b.vBuffer), cap(b.vBuffer), b.Available(), b.readIndex, b.writeIndex)
		return []byte{}
	}

	return b.vBuffer[b.readIndex : b.readIndex+int64(size)]
}

func (b *BufferIO) Peek16() uint16 {
	return BTUint16(b.PeekBytes(2))
}

func (b *BufferIO) Peek32() uint32 {
	return BTUint32(b.PeekBytes(4))
}

func (b *BufferIO) GetBytes(size int) (r []byte) {
	r = b.PeekBytes(size)
	if len(r) == 0 {
		return
	}

	b.Skip(size)
	return
}

func (b *BufferIO) GetUint8() uint8 {
	return BTUint8(b.GetBytes(1))
}

func (b *BufferIO) GetInt8() int8 {
	return BTInt8(b.GetBytes(1))
}

func (b *BufferIO) GetUint16() uint16 {
	return BTUint16(b.GetBytes(2))
}

func (b *BufferIO) GetInt16() int16 {
	return BTInt16(b.GetBytes(2))
}

func (b *BufferIO) GetUint32() uint32 {
	return BTUint32(b.GetBytes(4))
}

func (b *BufferIO) GetInt32() int32 {
	return BTInt32(b.GetBytes(4))
}

func (b *BufferIO) GetUint64() uint64 {
	return BTUint64(b.GetBytes(8))
}

func (b *BufferIO) GetInt64() int64 {
	return BTInt64(b.GetBytes(8))
}

func (b *BufferIO) GetFloat32() float32 {
	return BTFloat32(b.GetBytes(4))
}

func (b *BufferIO) GetFloat64() float64 {
	return BTFloat64(b.GetBytes(8))
}

func (b *BufferIO) GetString(size int) string {
	return string(b.GetBytes(size))
}

func (b *BufferIO) GetLine() (s string, ilen int) {
	s = ""
	ilen = -1
	pos := bytes.Index(b.vBuffer, []byte{'\n'})
	if pos == -1 {
		return
	}

	linePak := b.Slice(0, pos+1)
	linePak[pos] = 0
	s = string(linePak) // 得到这个字符串
	ilen = pos + 1
	return
}

/////////////////////////////////
// 写行为
func (b *BufferIO) WriteAuto(buf []byte) int {
	return b.Write(buf, len(buf))
}

func (b *BufferIO) Write(buf []byte, wlen int) int {
	endPos := b.writeIndex + int64(wlen)
	if endPos > b.Len() {
		err := b.Resize(endPos + defaultSize) // 扩容
		if err != nil {
			return -1
		}
	}

	if b.writeIndex > int64(len(b.vBuffer)) {
		fmt.Println(len(b.vBuffer), cap(b.vBuffer), b.Available(), b.readIndex, b.writeIndex)
		err := b.Resize(endPos + defaultSize) // 扩容
		if err != nil {
			return -1
		}
	}

	copy(b.vBuffer[b.writeIndex:], buf)
	b.writeIndex += int64(wlen)

	return wlen
}

func (b *BufferIO) WriteUInt8(v uint8) {
	b.WriteAuto(U8TBytes(v))
}

func (b *BufferIO) WriteInt8(v int8) {
	b.WriteAuto(I8TBytes(v))
}

func (b *BufferIO) WriteUInt16(v uint16) {
	b.WriteAuto(U16TBytes(v))
}

func (b *BufferIO) WriteInt16(v int16) {
	b.WriteAuto(I16TBytes(v))
}

func (b *BufferIO) WriteUInt32(v uint32) {
	b.WriteAuto(U32TBytes(v))
}

func (b *BufferIO) WriteInt32(v int32) {
	b.WriteAuto(I32TBytes(v))
}

func (b *BufferIO) WriteUInt64(v uint64) {
	b.WriteAuto(U64TBytes(v))
}

func (b *BufferIO) WriteInt64(v int64) {
	b.WriteAuto(I64TBytes(v))
}

func (b *BufferIO) WriteFloat32(v float32) {
	b.WriteAuto(F32TBytes(v))
}

func (b *BufferIO) WriteFloat64(v float64) {
	b.WriteAuto(F64TBytes(v))
}

func (b *BufferIO) WriteString(v string) {
	b.WriteAuto([]byte(v))
}

func (b *BufferIO) WriteLine(v string) {
	b.WriteAuto([]byte(v + "\n"))
}

/////////////////////////////
// 修改指定位置数据
func (b *BufferIO) ChangeAuto(pos int, buf []byte) {
	wlen := len(buf)
	if int64(pos+wlen) > b.Len() {
		return
	}

	for i := pos; i < pos+wlen; i++ {
		b.vBuffer[i] = buf[i-pos]
	}
}

func (b *BufferIO) ChangeUInt8(pos int, v uint8) {
	b.ChangeAuto(pos, U8TBytes(v))
}

func (b *BufferIO) ChangeInt8(pos int, v int8) {
	b.ChangeAuto(pos, I8TBytes(v))
}

func (b *BufferIO) ChangeUInt16(pos int, v uint16) {
	b.ChangeAuto(pos, U16TBytes(v))
}

func (b *BufferIO) ChangeInt16(pos int, v int16) {
	b.ChangeAuto(pos, I16TBytes(v))
}

func (b *BufferIO) ChangeUInt32(pos int, v uint32) {
	b.ChangeAuto(pos, U32TBytes(v))
}

func (b *BufferIO) ChangeInt32(pos int, v int32) {
	b.ChangeAuto(pos, I32TBytes(v))
}

func (b *BufferIO) ChangeUInt64(pos int, v uint64) {
	b.ChangeAuto(pos, U64TBytes(v))
}

func (b *BufferIO) ChangeInt64(pos int, v int64) {
	b.ChangeAuto(pos, I64TBytes(v))
}

func (b *BufferIO) ChangeFloat32(pos int, v float32) {
	b.ChangeAuto(pos, F32TBytes(v))
}

func (b *BufferIO) ChangeFloat64(pos int, v float64) {
	b.ChangeAuto(pos, F64TBytes(v))
}

func (b *BufferIO) ChangeString(pos int, v string) {
	b.ChangeAuto(pos, []byte(v))
}

// 通用解包函数
// 默认协议的首部4个字节为包长度，并返回一个仅有该数据内容的STREAM
// |len 4byte|body or header|
// Unpack后会自动shift
func (b *BufferIO) Unpack(fn CryptFn, key []byte) (buf []byte, err error) {
	buf = nil
	err = errors.New("Pack Unready...")
	if b.Available() < 4 {
		return // 包还没准备好
	}
	// 取出剩余长度
	len := int64(b.Peek32())
	if b.Available() < (len + 4) {
		return // 包还没准备好
	}

	b.Skip(4)                             // 跳过数据包长度
	bPak := fn(b.GetBytes(int(len)), key) // 读取body长度
	buf = make([]byte, (len + 4))
	copy(buf[:4], U32TBytes(uint32(len)))
	copy(buf[4:], bPak)
	err = nil

	return
}

// 解析以\n结尾的数据包并返回一个数据结构
func (b *BufferIO) UnpackLine(fn CryptFn, key []byte) (buf []byte, err error) {
	buf = nil
	err = errors.New("Pack Unready...")
	if b.Available() < 2 {
		return // 包还没准备好
	}

	sBuf, ilen := b.GetLine()
	if ilen == -1 {
		return // 不需要解压
	}

	b.Skip(ilen)
	bPak := []byte(sBuf) // 读取body长度
	buf = make([]byte, ilen-1)
	copy(buf, bPak[:ilen-1])
	err = nil

	return
}

// 通用打包体系
// |len 4byte|body or header|
func (b *BufferIO) Packer(fn CryptFn, key []byte) (buf []byte, err error) {
	buf = b.Bytes()
	err = errors.New("No Data Need Package...")
	if b.writeIndex < 4 {
		return
	}

	b.ChangeUInt32(0, uint32(b.writeIndex-4)) // 改变长
	data := b.Slice(0, int(b.writeIndex))
	unPak := fn(data[4:], key)

	// 修复加密错位
	buf = make([]byte, b.writeIndex)
	copy(buf[:4], data[:4])
	copy(buf[4:], unPak)
	err = nil
	return
}

func PackerOnce(in []byte, fn CryptFn, key []byte) (buf []byte, err error) {
	buffer := BuildPacker(in)
	buf, err = buffer.Packer(fn, key)
	return
}
