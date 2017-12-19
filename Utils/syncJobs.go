// 轻量级单对象并行运行池
// 会产生一定的数据COPY
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/12/18
package Utils

import (
	"sync"
)

type SyncJob struct {
	wait sync.WaitGroup
}

func (job *SyncJob) Append() error {
	return nil
}

func (job *SyncJob) StartWait() {
	job.wait.Wait()
}
