package grapeLogger

import (
	"testing"
)

func Test_Logger(t *testing.T) {
	BuildLogger("../log", "normal.log")

	INFO("testlogger,asdasd")
	ERROR("normal data")

	FLUSH()

	t.Log("success")
}

func Benchmark_write(t *testing.B) {
	BuildLogger("../test_log", "test.log")

	for i := 0; i < t.N; i++ {
		INFO("write_logger:%v", i)
		ERROR("write_error:%v", i)
	}

	FLUSH()
}

func Benchmark_Parallel(t *testing.B) {
	t.RunParallel(func(pb *testing.PB) {
		BuildLogger("../test_log", "test.log")
		for pb.Next() {
			INFO("write_logger")
			ERROR("write_error")
		}

		FLUSH()
	})
}
