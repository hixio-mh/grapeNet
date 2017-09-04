package grapeCSV

import (
	"fmt"
	"testing"
)

type testCsvNode struct {
	Name     string  `column:"name"`
	LastName string  `column:"tags"`
	Data     float32 `column:"data"`
	Value    int     `column:"dataval"`
}

func TestCSVTag(t *testing.T) {
	newCSV, err := NewCSVDefault("../_csv_tests/test002.csv")
	if err != nil {
		t.Error(err)
		return
	}

	sval := &testCsvNode{}
	newCSV.GetRow(0, sval)
	fmt.Println(sval)
}

// 测试自定义TAB为TOKEN的用例
func TestCSV_CustomTag(t *testing.T) {
	newCSV, err := NewCSV("../_csv_tests/test001.csv", '	', true)
	if err != nil {
		t.Error(err)
		return
	}

	sval := &testCsvNode{}
	newCSV.GetRow(0, sval)
	newCSV.CloseAll()
}

func TestCSV_EmptyFile(t *testing.T) {
	newCSV, err := CreateCSV("../_csv_tests/test003.csv", '	', testCsvNode{})
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 2000; i++ {
		newCSV.Append(testCsvNode{
			Name:     fmt.Sprintf("name:%d", i),
			LastName: fmt.Sprintf("temp:%d", i+1000),
			Data:     1.2222 + float32(i),
			Value:    i * 2000,
		})
	}

	newCSV.SaveAll()
	newCSV.CloseAll()
}

// 50W条数据只需要2S写出
func Benchmark_AppendFile(b *testing.B) {
	newCSV, err := CreateCSV("../_csv_tests/Benchmark001.csv", Default_token, testCsvNode{})
	if err != nil {
		b.Error(err)
		return
	}

	for i := 0; i < b.N; i++ {
		newCSV.Append(testCsvNode{
			Name:     fmt.Sprintf("name:%d", i),
			LastName: fmt.Sprintf("temp:%d", i+1000),
			Data:     1.2222 + float32(i),
			Value:    i * 2000,
		})
	}

	newCSV.SaveAll()
	newCSV.CloseAll()
}

// 100W数据 2S写出 线程安全
func Benchmark_Parallel(b *testing.B) {
	newCSV, err := CreateCSV("../_csv_tests/Benchmark002.csv", Default_token, testCsvNode{})
	if err != nil {
		b.Error(err)
		return
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			newCSV.Append(testCsvNode{
				Name:     fmt.Sprintf("name:%d", 1),
				LastName: fmt.Sprintf("temp:%d", 1000),
				Data:     1.2222,
				Value:    2000,
			})
		}
	})

	newCSV.SaveAll()
	newCSV.CloseAll()
}

// 50w数据解析仅仅需要1S
func TestCSV_ReadBenchmark(t *testing.T) {
	newCSV, err := NewCSVDefault("../_csv_tests/Benchmark001.csv")
	if err != nil {
		t.Error(err)
		return
	}

	sval := &testCsvNode{}
	newCSV.GetRow(0, sval)
	newCSV.CloseAll()
}
