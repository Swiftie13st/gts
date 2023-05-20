/**
  @author: Bruce
  @since: 2023/5/20
  @desc: //自旋锁
**/

package utils

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type spinLock uint32

// 反映锁竞争激烈程度，等待上限
const maxBackoff = 16

func (sl *spinLock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < maxBackoff {
			// 每次等待翻倍
			backoff <<= 1
		}
	}
}

func (sl *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func NewSpinLock() sync.Locker {
	return new(spinLock)
}
