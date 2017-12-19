// 一款带锁的列表
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/12/10
package continer

import (
	"container/list"
	"sync"
)

type SList struct {
	slist  *list.List
	locker sync.RWMutex
}

func New() *SList {
	return &SList{
		slist: list.New(),
	}
}

func (sc *SList) Push(item interface{}) {
	sc.locker.Lock()
	defer sc.locker.Unlock()

	sc.slist.PushBack(item)
}

func (sc *SList) First() interface{} {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	return sc.slist.Front()
}

func (sc *SList) Back() interface{} {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	return sc.slist.Back()
}

func (sc *SList) Clear() {
	sc.locker.Lock()
	defer sc.locker.Unlock()

	sc.slist = list.New() //
}

func (sc *SList) Range(fn func(i interface{})) {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	for e := sc.slist.Front(); e != nil; e = e.Next() {
		fn(e.Value)
	}
}

func (sc *SList) ReverseRange(fn func(i interface{})) {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	for e := sc.slist.Back(); e != nil; e = e.Prev() {
		fn(e.Value)
	}
}

func (sc *SList) Search(fn func(i interface{}) bool) (interface{}, bool) {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	for e := sc.slist.Front(); e != nil; e = e.Next() {
		if fn(e.Value) {
			return e.Value, true
		}
	}

	return nil, false
}

func (sc *SList) Remove(fn func(i interface{}) bool) {
	sc.locker.Lock()
	defer sc.locker.Unlock()

	for e := sc.slist.Front(); e != nil; e = e.Next() {
		if fn(e.Value) {
			sc.slist.Remove(e)
			return
		}
	}
}
