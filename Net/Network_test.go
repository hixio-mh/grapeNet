package grapeNet

import (
	"testing"
)

func Test_Listens(t *testing.T) {
	newTcp, err := NewTcpServer(":9234")
	if err != nil {
		t.Error(err)
		return
	}

	newTcp.Runnable()
}
