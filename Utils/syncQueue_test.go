package Utils

import (
	"fmt"
	"testing"
)

var queue = NewSQueue()

func procDequque() {
	for {
		v := queue.Pop()
		fmt.Println(v)
	}
}

func Benchmark_syncQueue(b *testing.B) {
	go procDequque()
	go procDequque()
	go procDequque()
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				queue.Push("123123123123")
				queue.Push("123123123123")
				queue.Push("123123123123")
			}
		})

}

func Benchmark_Dequeue(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				queue.Push("123123123123")
				queue.Pop()
			}
		})

}
