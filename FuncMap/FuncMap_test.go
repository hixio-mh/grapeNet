package grapeFunc

import (
	"fmt"
	"reflect"
	"testing"
)

func vestAbc(s string, i int) {
	fmt.Println(s, i)
}

func vest3(i uint32, s string, data string) {
	//fmt.Println(i, s, data)
}

func vest4Result(i uint32, s string, data string) (uint32, string) {
	return i, s
}

type iTest interface {
	Write()
}

type TestIns struct {
}

func (c *TestIns) Write() {
	fmt.Println("TestIns write...")
}

func vestInterface(i interface{}, test iTest) {
	fmt.Println(i)
	test.Write()
}

func Test_InterfaceCall(t *testing.T) {

	reflect.TypeOf(new(TestIns))

	FastBind("interface", vestInterface)
	err := FastCall("interface", 3000, new(TestIns))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_MapCall(t *testing.T) {
	FastBind("0", vestAbc)
	FastBind(1, vestAbc)

	FastBind(2.0, vest3)
	FastBind("CCCC", vest3)

	FastBind("0Result", vest4Result)

	FastCall("0", "Call 0", 1233)
	FastCall("0", "Call 0 Sc", 1233, "asdasd", 4444)
	FastCall(1, "Call 1", 2000)

	FastCall("CCCC", uint32(2000), "asdasd", "zxxczxcxc")
	FastCall(2.0, uint32(3000), "Call_Float", "zxxczxcxc")

	res, err := FastCallR("0Result", uint32(3000), "Call_Float", "zxxczxcxc")
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	fmt.Println(res)
}

func Benchmark_MapCall(t *testing.B) {
	for i := 0; i < t.N; i++ {
		FastBind(i, vest3)
	}
}

func Benchmark_CallBM(t *testing.B) {
	FastBind("ABC", vest3)

	for i := 0; i < t.N; i++ {
		FastCall("ABC", uint32(i), "asdasd", "zxzxcasd")
	}
}

func Benchmark_Parallel(b *testing.B) {
	FastBind("ABC", vest3)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 500; i++ {
				FastCall("ABC", uint32(i), "asdasd", "Zzxcaasd")
			}
		}
	})
}
