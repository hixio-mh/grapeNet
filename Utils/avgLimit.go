package Utils

import (
	"fmt"
	"sync"
)

// 平均数值同步计算容器
type AVGLimit struct {
	total float64
	avg   float64
	count float64
	lock  sync.RWMutex
}

func (c *AVGLimit) Add(value float64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.total += value
	c.count += 1
	c.avg = c.total / c.count
}

func (c *AVGLimit) Reset() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.total = 0.0
	c.count = 0.0
	c.avg = 0.0
}

func (c *AVGLimit) String() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return fmt.Sprintf("total:%.2f avg:%.2f count:%.2f\n",
		c.total, c.avg, c.count)
}

func (c *AVGLimit) Value() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.avg
}
