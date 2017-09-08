package grapeSign

import (
	"testing"
	"time"
)

type SignTest struct {
	TestAbc string `form:"abc" json:"jabc" sign:"abc"`
	DataAbc int    `form:"tint" json:"jtint" sign:"tint"`
	Time    int64  `form:"t" json:"jt" sign:"t"`
	Sign    string `form:"sign" b:"-" sign:"-"`
}

func TestSignForm(t *testing.T) {
	SignTag = "sign"
	SignKey = "20f0d253d40714277e5c12081db1237cafdc3999"

	st := &SignTest{
		TestAbc: "123123asdasd",
		DataAbc: 30000,
		Time:    time.Now().Unix(),
	}

	sign, serr := SignMD5(st)
	if serr != nil {
		t.Error(serr)
		return
	}

	st.Sign = sign
	t.Log(sign)
}

func TestSignJson(t *testing.T) {
	SignTag = "json"
	SignKey = "20f0d253d40714277e5c12081db1237cafdc3999"
	IsSort = false

	st := &SignTest{
		TestAbc: "123123asdasd",
		DataAbc: 30000,
		Time:    time.Now().Unix(),
	}

	sign, serr := SignMD5(st)
	if serr != nil {
		t.Error(serr)
		return
	}

	st.Sign = sign
	t.Log(sign)
}

func TestSignMap(t *testing.T) {
	SignKey = "20f0d253d40714277e5c12081db1237cafdc3999"

	st := &SignTest{
		TestAbc: "123123asdasd",
		DataAbc: 30000,
		Time:    time.Now().Unix(),
	}

	stMap, serrv := Type2Map(st)
	if serrv != nil {
		t.Error(serrv)
		return
	}

	sign, serr := SignSha1(stMap)
	if serr != nil {
		t.Error(serr)
		return
	}

	st.Sign = sign
	t.Log(sign)
}

func Benchmark_Count(b *testing.B) {
	st := &SignTest{
		TestAbc: "123123asdasd",
		DataAbc: 30000,
		Time:    time.Now().Unix(),
	}

	for i := 0; i < b.N; i++ {
		sign, serr := SignMD5(st)
		if serr != nil {
			b.Error(serr)
			return
		}

		st.Sign = sign
	}

}

func Benchmark_ParallelsCount(b *testing.B) {
	st := &SignTest{
		TestAbc: "123123asdasd",
		DataAbc: 30000,
		Time:    time.Now().Unix(),
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sign, serr := SignMD5(st)
			if serr != nil {
				b.Error(serr)
				return
			}

			st.Sign = sign
		}
	})
}
