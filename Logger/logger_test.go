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
