// 针对Interface转换为指定类型，内部检测并返回默认数值
// 强制类型转换，非可转换类型直接报错
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/12/14

package Utils

import (
	"fmt"
	"reflect"
	"strconv"
)

var (
	vaildKind map[reflect.Kind]bool = map[reflect.Kind]bool{
		reflect.String: true,
		reflect.Bool:   true,
		reflect.Int:    true, reflect.Int16: true, reflect.Int32: true, reflect.Int64: true,
		reflect.Uint: true, reflect.Uint16: true, reflect.Uint32: true, reflect.Uint64: true,
		reflect.Int8: true, reflect.Float32: true, reflect.Float64: true,
	}
)

func MustString(v interface{}) string {
	result, ok := v.(string)
	if !ok {
		return fmt.Sprint(v)
	}

	return result
}

func isVaildKind(kind, target reflect.Kind) bool {
	// 最佳情况，类型相等
	if kind == target {
		return true
	}

	if kind == reflect.Bool && target != reflect.String {
		return false
	}

	// 所有ptr,struct,map,byte等全部禁止转换
	_, ok := vaildKind[kind]
	if !ok {
		return false
	}

	_, ok = vaildKind[target]
	if !ok {
		return false
	}

	return true
}

func convertValue(v reflect.Value, target reflect.Type) (rv reflect.Value, err error) {
	err = nil
	rv = reflect.Value{} // empty value
	if isVaildKind(v.Kind(), target.Kind()) == false {
		err = fmt.Errorf("type error,convert faild...")
		return
	}

	// 来源如果是string则单独处理
	if v.Kind() == reflect.String {
		var iv reflect.Value
		switch target.Kind() {
		case reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			i, verr := strconv.ParseInt(v.String(), 10, 64)
			if verr != nil {
				err = verr
				return
			}

			iv = reflect.ValueOf(i)
		case reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			i, verr := strconv.ParseUint(v.String(), 10, 64)
			if verr != nil {
				err = verr
				return
			}

			iv = reflect.ValueOf(i)
		case reflect.Float32, reflect.Float64:
			i, verr := strconv.ParseFloat(v.String(), 10)
			if verr != nil {
				err = verr
				return
			}

			iv = reflect.ValueOf(i)
		case reflect.Bool:
			i, verr := strconv.ParseBool(v.String())
			if verr != nil {
				err = verr
				return
			}

			rv = reflect.ValueOf(i)
			return
		}

		rv = iv.Convert(target)
	} else {
		rv = v.Convert(target)
	}

	return
}

func MustBool(v interface{}, must bool) bool {
	rv, err := convertValue(reflect.ValueOf(v), reflect.TypeOf(must))
	if err != nil {
		return must
	}

	return rv.Bool()
}

func MustInt(v interface{}, must int) int {
	rv, err := convertValue(reflect.ValueOf(v), reflect.TypeOf(must))
	if err != nil {
		return must
	}

	return int(rv.Int())
}

func MustInt64(v interface{}, must int64) int64 {
	rv, err := convertValue(reflect.ValueOf(v), reflect.TypeOf(must))
	if err != nil {
		return must
	}

	return rv.Int()
}

func MustUInt64(v interface{}, must uint64) uint64 {
	rv, err := convertValue(reflect.ValueOf(v), reflect.TypeOf(must))
	if err != nil {
		return must
	}

	return rv.Uint()
}

func MustFloat64(v interface{}, must float64) float64 {
	rv, err := convertValue(reflect.ValueOf(v), reflect.TypeOf(must))
	if err != nil {
		return must
	}

	return rv.Float()
}
