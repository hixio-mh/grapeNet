// 字符串拆分并转换为指定数据
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/24

package grapeStream

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultToken = " "
)

type StrLine struct {
	source string

	command []string

	nextIdx int

	Token string
}

func NewSL(s string) *StrLine {
	sl := &StrLine{
		source:  s,
		command: []string{},
		nextIdx: 0,
		Token:   defaultToken,
	}

	sl.Parser()
	return sl
}

func NewSLEmpty() *StrLine {
	return &StrLine{
		source:  "",
		command: []string{},
		nextIdx: 0,
		Token:   defaultToken,
	}
}

func (s *StrLine) Source() string {
	return s.source
}

func (s *StrLine) Parser() {
	s.command = strings.Split(s.source, s.Token) // 拆分数据
}

func (s *StrLine) Pack() string {
	s.source = strings.Join(s.command, s.Token)
	return s.source
}

func (s *StrLine) Command() string {
	return s.Get(0)
}

/// 获取数据
func (s *StrLine) Get(args int) string {
	if len(s.command) <= args {
		return ""
	}

	return s.command[args]
}

func (s *StrLine) GetInt(args int) int {
	iv, _ := strconv.Atoi(s.Get(args))
	return iv
}

func (s *StrLine) GetNext() string {
	if len(s.command) <= s.nextIdx {
		return ""
	}

	ss := s.command[s.nextIdx]
	s.nextIdx++
	return ss
}

func (s *StrLine) GetNextInt() int {
	iv, _ := strconv.Atoi(s.GetNext())
	return iv
}

/// 写入数据
func (s *StrLine) CreateCmd(sCmd string) {
	s.command = []string{}
	s.command = append(s.command, sCmd)
}

func (s *StrLine) Append(args interface{}) {
	s.command = append(s.command, fmt.Sprintf("%v", args))
}

func (s *StrLine) AppendA62(args int) {
	s.command = append(s.command, CNV10to62(args))
}
