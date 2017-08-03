package grapeFunc

import (
	"fmt"
	"testing"
)

func vestAbc(s string, i int) {
	fmt.Println(s, i)
}

func vest3(i uint32, s string, data string) {
	//fmt.Println(i, s, data)
}

func Test_MapCall(t *testing.T) {
	FastBind("0", vestAbc)
	FastBind(1, vestAbc)

	FastBind(2.0, vest3)
	FastBind("CCCC", vest3)

	FastCall("0", "Call 0", 1233)
	FastCall("0", "Call 0 Sc", 1233, "asdasd", 4444)
	FastCall(1, "Call 1", 2000)

	FastCall("CCCC", uint32(2000), "asdasd", "zxxczxcxc")
	FastCall(2.0, uint32(3000), "Call_Float", "zxxczxcxc")

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
