// golang 三元运算符的合并版本
// 由于golang没有三元运算符所以补充一套
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/12/10

package Utils

// 当cond为true 返回a 否则返回b
// 支持a,b为任意不同的类型
func If(cond bool, a interface{}, b interface{}) interface{} {
	if cond {
		return a
	}

	return b
}

// 日常类通用模块
func Ifs(cond bool, a string, b string) string {
	return If(cond, a, b).(string)
}

func Ifn(cond bool, a int, b int) int {
	return If(cond, a, b).(int)
}

func Ifn64(cond bool, a int, b int64) int64 {
	return If(cond, a, b).(int64)
}

func Ifd(cond bool, a float64, b float64) float64 {
	return If(cond, a, b).(float64)
}
