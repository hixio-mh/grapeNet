package grapeConn

import (
	"testing"
)

func Test_Conn(t *testing.T) {
	NECM := NewCM()

	newCnn := &Conn{
		SessionId: CreateUUID(1),
	}

	NECM.Register <- newCnn
}

func Benchmark_Register(b *testing.B) {
	NECM := NewCM()

	sessionIds := []string{}

	for i := 0; i < b.N; i++ {
		newCnn := &Conn{
			SessionId: CreateUUID(1),
		}

		NECM.Register <- newCnn

		sessionIds = append(sessionIds, newCnn.GetSessionId())
	}

	for i := 0; i < len(sessionIds); i++ {
		nc := NECM.Get(sessionIds[i])
		if nc == nil {
			continue
		}

		NECM.Unregister <- nc
	}
}

func Benchmark_RegParallel(b *testing.B) {
	NECM := NewCM()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			newCnn := &Conn{
				SessionId: CreateUUID(1),
			}

			NECM.Register <- newCnn
		}
	})
}
