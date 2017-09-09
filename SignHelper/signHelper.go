// 直接根据类型生成指定的MD5或SHA1的SIGN
// 用于API或任何参数校验,时间参数t建议自己加在结构中
// 基本算法为 value组合+_+key 或 key=value组合+_+key
// 建议传入的是一个结构体，不要是map之类的，因为map是无序的，如果传入MAP自动按照名称排序一次
// 实现自动化以后可以少写好多代码并且可以验证一个包或一个API的参数是否合法
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/09/08
package grapeSign

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

var IsUseKey bool = true      // 是否只使用数值组合生成 [key=value]
var IsSort bool = true        // 是否对key做字母排序，默认开启排序，可不排序直接生成
var SignKey string = ""       // 一个默认的KEY
var SignTag string = "sign"   // 可以根据json或form等来生成sign,设置为 - 时代表不做任何生成行为
var SignSplitTag string = "_" // 默认下划线分割
var SignMergeTag string = "&" // join时用的数据，可以自行设置

func SortMap2Str(tmap map[string]interface{}) string {
	var keySort []string
	for k, _ := range tmap {
		keySort = append(keySort, k)
	}

	if IsSort {
		sort.Strings(keySort) // 升序排序KEY
	}

	signArgs := []string{}
	for _, vk := range keySort {
		if IsUseKey {
			signArgs = append(signArgs, fmt.Sprintf("%v=%v", vk, tmap[vk]))
		} else {
			signArgs = append(signArgs, fmt.Sprintf("%v", vk, tmap[vk]))
		}
	}

	return strings.Join(signArgs, SignMergeTag)
}

func Type2Map(t interface{}) (tmap map[string]interface{}, err error) {
	tmap = make(map[string]interface{})
	err = nil
	v := reflect.ValueOf(t)

	if v.Kind() == reflect.Map {
		// 得到map的keys
		keys := v.MapKeys()

		// 赋值给Map
		for _, vk := range keys {
			if vk.Kind() != reflect.String {
				continue
			}

			tmap[vk.String()] = v.MapIndex(vk).Interface()
		}
		return
	} else if v.Kind() == reflect.Ptr || v.Kind() == reflect.Struct {
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			err = errors.New("reflect type must be struct or struct ptr...")
			fmt.Println(err)
			return
		}

		// 反射并将类型转换为map
		t := v.Type() // 得到反射类型
		// 根据结构内的数据 开始进行反射
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			vf := v.Field(i)

			if sf.Tag.Get(SignTag) == "-" {
				continue // 该类型不进行任何的操作
			}

			if vf.CanSet() == false {
				continue // 未公开不可
			}

			if sf.Tag.Get("b") == "-" {
				continue // 特殊屏蔽
			}

			keyName := sf.Tag.Get(SignTag)
			// 此处没有tag那么尝试一下名字
			if len(keyName) == 0 {
				keyName = sf.Name
				if len(keyName) == 0 {
					continue // 还是没有那么不设置
				}
			}

			tmap[keyName] = vf.Interface()
		}
	} else {
		err = errors.New("reflect type must be struct or struct ptr...")
		fmt.Println(err)
		return
	}

	return
}

func SignMD5NE(t interface{}) string {
	sign, _ := KeySignMD5(t, SignKey)
	return sign
}

func SignSha1NE(t interface{}) string {
	sign, _ := KeySignSha1(t, SignKey)
	return sign
}

func KeySignMD5NE(t interface{}, key string) string {
	sign, _ := KeySignMD5(t, key)
	return sign
}

func KeySignSha1NE(t interface{}, key string) string {
	sign, _ := KeySignSha1(t, key)
	return sign
}

func SignMD5(t interface{}) (sign string, err error) {
	return KeySignMD5(t, SignKey)
}

func SignSha1(t interface{}) (sign string, err error) {
	return KeySignSha1(t, SignKey)
}

func KeySignMD5(t interface{}, key string) (sign string, err error) {
	sign = ""
	err = nil

	tmap, terr := Type2Map(t)
	if terr != nil {
		err = terr
		return
	}

	md5hash := md5.New()
	md5hash.Write([]byte(SortMap2Str(tmap)))
	sign = hex.EncodeToString(md5hash.Sum(nil))
	return
}

func KeySignSha1(t interface{}, key string) (sign string, err error) {
	sign = ""
	err = nil

	tmap, terr := Type2Map(t)
	if terr != nil {
		err = terr
		return
	}

	sha1hash := sha1.New()
	sha1hash.Write([]byte(SortMap2Str(tmap)))
	sign = hex.EncodeToString(sha1hash.Sum(nil))
	return
}
