/**
  @author: Bruce
  @since: 2023/5/21
  @desc: //TODO
**/

package pool

import "time"

type workerStack struct {
	items  []worker
	expiry []worker
}

func newWorkerStack(size int) *workerStack {
	return &workerStack{
		items: make([]worker, 0, size),
	}
}

// 当前数量
func (wq *workerStack) len() int {
	return len(wq.items)
}

// 是否为空
func (wq *workerStack) isEmpty() bool {
	return len(wq.items) == 0
}

// 插入一个worker
func (wq *workerStack) insert(w worker) error {
	wq.items = append(wq.items, w)
	return nil
}

// 获取一个worker
func (wq *workerStack) detach() worker {
	l := wq.len()
	if l == 0 {
		return nil
	}

	w := wq.items[l-1]
	wq.items[l-1] = nil // avoid memory leaks
	wq.items = wq.items[:l-1]

	return w
}

// 回收等待时间长的worker
func (wq *workerStack) refresh(duration time.Duration) []worker {
	n := wq.len()
	if n == 0 {
		return nil
	}

	expiryTime := time.Now().Add(-duration)
	// 通过二分找到对应时间的下标
	index := wq.binarySearch(0, n-1, expiryTime)

	wq.expiry = wq.expiry[:0]
	if index != -1 {
		wq.expiry = append(wq.expiry, wq.items[:index+1]...)
		m := copy(wq.items, wq.items[index+1:])
		for i := m; i < n; i++ {
			wq.items[i] = nil
		}
		wq.items = wq.items[:m]
	}
	return wq.expiry
}

func (wq *workerStack) binarySearch(l, r int, expiryTime time.Time) int {
	var mid int
	for l < r {
		mid = (l + r) >> 1
		if expiryTime.Before(wq.items[mid].lastUsedTime()) {
			r = mid
		} else {
			l = mid + 1
		}
	}
	return r
}

// 重置
func (wq *workerStack) reset() {
	for i := 0; i < wq.len(); i++ {
		wq.items[i].finish()
		wq.items[i] = nil
	}
	wq.items = wq.items[:0]
}
