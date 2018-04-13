// 部分位置用于取代CHAN的无限无阻塞队列
// version 1.0 beta
// 需要特别的LOCKFREE队列 ，尝试 github.com/yireyun/go-queue
// by koangel
// email: jackliu100@gmail.com
// 2018/04/13
package Utils

import (
	"sync"

	qv1 "gopkg.in/eapache/queue.v1"
)

type SyncQueue struct {
	cond   *sync.Cond
	l      *qv1.Queue
	locker sync.Mutex
}

func NewSQueue() *SyncQueue {
	q := &SyncQueue{
		l: qv1.New(),
	}

	q.cond = sync.NewCond(&q.locker)
	return q
}

func (q *SyncQueue) Push(value interface{}) {
	q.locker.Lock()
	defer q.locker.Unlock()

	q.l.Add(value)
	q.cond.Signal()
}

func (q *SyncQueue) Pop() (val interface{}) {
	q.locker.Lock()
	defer q.locker.Unlock()

	// WAIT内部会有锁释放，所以这里不会造成死锁，当队列为空时，进行等待唤醒
	for q.l.Length() <= 0 {
		q.cond.Wait()
	}

	val = q.l.Remove()

	return
}
