// 协议的接口类
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/13

package grapeProto

type ProtoInter interface {
	PackData(val interface{}) []byte
	UnpackData(data []byte,len int) interface{}
}
