// 自动反射和解析CSV文件，对于有列的文件则通过列名解析
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/8/16
package grapeCSV

import (
	"encoding/csv"
	"errors"
	"os"
	"reflect"
	"strconv"
	"sync"
)

const (
	Default_token   = ','
	Default_comment = '#'

	tag_column = "column"
	tag_index  = "colIdx"
)

type ParserCSV struct {
	isOpen     bool
	isSkipHead bool
	// Data
	data [][]string
	// Header
	headers []string
	// col len
	columnNum int

	// reader
	reader *csv.Reader
	// writer
	writer *csv.Writer

	// Files
	files *os.File

	// once
	once sync.Once

	// 读写锁
	rwlocker sync.RWMutex
}

func NewCSVDefault(filename string) (csv *ParserCSV, err error) {
	csv = &ParserCSV{
		isOpen:     false,
		isSkipHead: false,
		columnNum:  0,
		reader:     nil,
		writer:     nil,
	}

	err = csv.OpenDefault(filename)
	return
}

func NewCSV(filename string, token rune, skipHead bool) (csv *ParserCSV, err error) {
	csv = &ParserCSV{
		isOpen:     false,
		isSkipHead: skipHead,
		columnNum:  0,
		reader:     nil,
		writer:     nil,
	}

	err = csv.Open(filename, token, skipHead)
	return
}

func CreateCSV(filename string, token rune, header interface{}) (csv *ParserCSV, err error) {
	csv = &ParserCSV{
		isOpen:     false,
		isSkipHead: false,
		columnNum:  0,
		reader:     nil,
		writer:     nil,
	}

	err = csv.Create(filename, token)
	if err != nil {
		return
	}

	csv.SetHeader(header)
	return
}

func (c *ParserCSV) Create(filename string, token rune) error {
	csvfile, err := os.Create(filename)
	if err != nil {
		return err
	}

	c.files = csvfile
	c.isOpen = true
	c.reader = csv.NewReader(csvfile)
	c.writer = csv.NewWriter(csvfile)

	c.reader.Comma = token
	c.reader.Comment = Default_comment

	c.writer.Comma = token

	return nil
}

func (c *ParserCSV) OpenDefault(filename string) error {
	return c.Open(filename, Default_token, true)
}

func (c *ParserCSV) Open(filename string, token rune, skipHead bool) error {
	csvfile, err := os.Open(filename)
	if err != nil {
		return err
	}

	c.files = csvfile
	c.isOpen = true
	c.reader = csv.NewReader(csvfile)
	c.writer = csv.NewWriter(csvfile)

	c.reader.Comma = token
	c.reader.Comment = Default_comment

	c.writer.Comma = token

	pData, rderr := c.reader.ReadAll()
	if rderr != nil {
		return rderr
	}

	c.isSkipHead = skipHead
	c.rwlocker.Lock()
	if len(pData) <= 0 {
		pData = [][]string{[]string{}}
		c.columnNum = 0
	} else {
		if skipHead {
			c.headers = pData[0] // 读取0行
			if len(pData) > 1 {
				c.data = pData[1:] //剩下的读进去
			} else {
				c.data = [][]string{[]string{}}
			}
		} else {
			c.headers = []string{}
			c.data = pData
		}

		c.columnNum = len(pData[0])
	}
	c.rwlocker.Unlock()

	return nil
}

func (c *ParserCSV) CloseAll() {
	c.once.Do(func() {
		if c.isOpen == false {
			return
		}

		c.isOpen = false
		c.writer.Flush()

		c.files.Close()
		c.reader = nil
		c.writer = nil

		c.headers = nil
		c.data = nil
	})
}

func (c *ParserCSV) SaveAll() {
	if c.isOpen == false {
		return
	}

	if c.isSkipHead {
		c.writer.Write(c.headers)
	}

	c.writer.WriteAll(c.data)
	c.writer.Flush()
}

func (c *ParserCSV) RowCount() int {
	return len(c.data)
}

func (c *ParserCSV) GetRow(row int, val interface{}) error {
	if c.isOpen == false {
		return errors.New("file not exists...")
	}

	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()
	if row > len(c.data) {
		return errors.New("row index overflow...")
	}

	return c.serialize(c.data[row], val)
}

func (c *ParserCSV) SetHeader(val interface{}) {
	if c.isOpen == false {
		return
	}

	v := reflect.ValueOf(val)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	newHeader := []string{}
	c.isSkipHead = true

	// 开始反射
	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		newHeader = append(newHeader, structField.Tag.Get(tag_column))
	}

	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()
	c.headers = newHeader
	c.columnNum = len(newHeader)
}

func (c *ParserCSV) SetRow(row int, val interface{}) error {
	if c.isOpen == false {
		return errors.New("file not exists...")
	}

	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()

	if row > len(c.data) {
		return errors.New("row index overflow...")
	}

	strData, derr := c.deserialize(val)
	if derr != nil {
		return derr
	}

	c.data[row] = strData
	return nil
}

func (c *ParserCSV) Append(val interface{}) error {
	if c.isOpen == false {
		return errors.New("file not exists...")
	}

	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()

	strData, derr := c.deserialize(val)
	if derr != nil {
		return derr
	}

	c.data = append(c.data, strData)
	return nil
}

////////////////////////////////////////////////////////////////
// tag转换为index
func (c *ParserCSV) getHeadCol(tag reflect.StructTag) int {
	coldata := tag.Get(tag_column)
	if c.isSkipHead && coldata != "" {
		for i := 0; i < len(c.headers); i++ {
			if c.headers[i] == coldata {
				return i
			}
		}

		return -1
	}

	colIdx, _ := strconv.Atoi(tag.Get(tag_index))
	if colIdx < 0 || colIdx > c.columnNum {
		return -1
	}

	return colIdx
}

func (c *ParserCSV) serialize(src []string, val interface{}) error {
	v := reflect.ValueOf(val)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("reflect type must be struct or struct ptr...")
	}

	t := v.Type()

	// 开始反射
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)

		if !f.CanSet() {
			continue
		}

		colIdx := c.getHeadCol(structField.Tag)
		if colIdx == -1 {
			continue // 错误不解析
		}

		value := c.formatValue(&f, src[colIdx])
		f.Set(value)
	}

	return nil
}

func (c *ParserCSV) formatKind(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Invalid:
		return "invalid"
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.String:
		return v.String()
	default: // reflect.Array, reflect.Struct, reflect.Interface
		return v.Type().String() + " value"
	}
}

func (c *ParserCSV) formatValue(v *reflect.Value, src string) reflect.Value {
	switch v.Kind() {
	case reflect.Invalid:
		return reflect.ValueOf("invalid")
	case reflect.Int, reflect.Int32, reflect.Int64,
		reflect.Int8, reflect.Int16:
		i, _ := strconv.ParseInt(src, 10, 32)
		return reflect.ValueOf(i).Convert(v.Type())
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		i, _ := strconv.ParseUint(src, 10, 32)
		return reflect.ValueOf(i).Convert(v.Type())
	case reflect.Bool:
		i, _ := strconv.ParseBool(src)
		return reflect.ValueOf(i)
	case reflect.Float32, reflect.Float64:
		i, _ := strconv.ParseFloat(src, 32)
		return reflect.ValueOf(i).Convert(v.Type())
	case reflect.String:
		return reflect.ValueOf(src)
	default: // reflect.Array, reflect.Struct, reflect.Interface
		return reflect.ValueOf(v.Type().String() + " value")
	}
}

func (c *ParserCSV) deserialize(val interface{}) (ref []string, err error) {
	v := reflect.ValueOf(val)

	err = nil
	ref = []string{}

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		err = errors.New("reflect type must be struct or struct ptr...")
		return
	}

	t := v.Type()
	// 构建一个空的结构
	for ic := 0; ic < c.columnNum; ic++ {
		ref = append(ref, "")
	}

	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)

		colIdx := c.getHeadCol(structField.Tag)
		if colIdx == -1 {
			continue // 错误不解析
		}

		ref[colIdx] = c.formatKind(f)
	}

	return
}
