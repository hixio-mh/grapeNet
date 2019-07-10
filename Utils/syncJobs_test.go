package Utils

import (
	"fmt"
	"strings"
	"testing"
)

func Test_JobsAppend(t *testing.T) {
	jobs := &SyncJob{}

	err := jobs.Append(func(a string) {
		fmt.Println(a, "inter call")
	}, "args1")
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	err = jobs.Append(func(a string) {
		fmt.Println(a, "inter call 02")
	}, "args2")
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	jobs.StartWait()
}

func Test_JobsAppendR(t *testing.T) {
	jobs := &SyncJob{}

	err := jobs.AppendR(func(a, rb string) string {
		fmt.Println(a, "inter call")
		return rb
	}, func(r string) {
		fmt.Println(r, "return")
	}, "args1", "return")
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	err = jobs.AppendR(func(a, rb string) string {
		fmt.Println(a, "inter call 02")
		return rb
	}, func(r string) {
		fmt.Println(r, "return 02")
	}, "args2", "return2")
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	jobs.StartWait()
}

func Test_EmptyJobs(t *testing.T) {
	jobs := &SyncJob{}
	jobs.StartWait()
}

func Test_SliceJobs(t *testing.T) {
	jobStr := "a,b,c,d,e,f,g,a,c,asd,a,a,a,a,s,s,s,d,d,a,a,sd,d,a,s"
	sliceStr := strings.Split(jobStr, ",")
	jobs := &SyncJob{}
	jobs.SliceJob(sliceStr, 2, func(start, end int) {
		fmt.Println(start, end, sliceStr[start:end])
		for i := start; i < end; i++ {
			//分段处理
		}
	})
	jobs.StartWait()
}

func Benchmark_Jobs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jobs := &SyncJob{}

		err := jobs.AppendR(func(a, rb string) string {
			//fmt.Println(a, "inter call")
			return rb
		}, func(r string) {
			//fmt.Println(r, "return")
		}, "args1", "return")
		if err != nil {
			fmt.Println(err)
			b.Fail()
			return
		}

		err = jobs.AppendR(func(a, rb string) string {
			//fmt.Println(a, "inter call 02")
			return rb
		}, func(r string) {
			//fmt.Println(r, "return 02")
		}, "args2", "return2")
		if err != nil {
			fmt.Println(err)
			b.Fail()
			return
		}

		jobs.StartWait()
	}
}
