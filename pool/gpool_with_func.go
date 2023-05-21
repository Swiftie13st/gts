/**
  @author: Bruce
  @since: 2023/5/21
  @desc: //有传参函数的pool
**/

package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type GPoolWithFunc struct {
	lock sync.Locker
	// 上限容量
	cap int32
	// 当前正在运行的goroutine
	running int32
	// 协程池的状态
	state int32
	// 并发协调器
	cond *sync.Cond
	// 协程对象池
	workerCache sync.Pool
	// 可用的协程列表
	workers workerQueue
	// 当前等待的数量
	waiting int32
	// 任务函数
	poolFunc func(interface{})
}

// 获取一个worker
func (p *GPoolWithFunc) retrieveWorker() (w worker) {
	spawnWorker := func() {
		w = p.workerCache.Get().(*goWorkerWithFunc)
		w.run()
	}

	p.lock.Lock()
	w = p.workers.detach()
	if w != nil {
		p.lock.Unlock()
	} else if capacity := p.Cap(); capacity == -1 || capacity > p.Running() {
		// 创建一个新worker
		p.lock.Unlock()
		spawnWorker()
	} else {
		// 已达上限，等待有worker放回
	retry:

		p.addWaiting(1)
		p.cond.Wait() // 阻塞等待
		p.addWaiting(-1)

		if p.IsClosed() {
			p.lock.Unlock()
			return
		}

		if w = p.workers.detach(); w == nil {
			if p.Free() > 0 {
				p.lock.Unlock()
				spawnWorker()
				return
			}
			goto retry
		}
		p.lock.Unlock()
	}
	return
}

// 放回一个worker
func (p *GPoolWithFunc) revertWorker(worker *goWorkerWithFunc) bool {
	if capacity := p.Cap(); (capacity > 0 && p.Running() > capacity) || p.IsClosed() {
		p.cond.Broadcast()
		return false
	}

	worker.lastUsed = time.Now()

	p.lock.Lock()
	if p.IsClosed() {
		p.lock.Unlock()
		return false
	}
	if err := p.workers.insert(worker); err != nil {
		p.lock.Unlock()
		return false
	}
	p.cond.Signal()
	p.lock.Unlock()

	return true
}

func NewPoolWithFunc(size int, pf func(interface{})) (*GPoolWithFunc, error) {
	if size <= 0 {
		size = -1
	}

	p := &GPoolWithFunc{
		cap:      int32(size),
		lock:     NewSpinLock(),
		poolFunc: pf,
	}
	p.workerCache.New = func() interface{} {
		return &goWorkerWithFunc{
			pool: p,
			args: make(chan interface{}, 1),
		}
	}
	p.cond = sync.NewCond(p.lock)
	p.workers = newWorkerStack(size)
	return p, nil
}

func (p *GPoolWithFunc) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == 1
}

// Waiting 当前正在等待运行的goroutine数量
func (p *GPoolWithFunc) Waiting() int {
	return int(atomic.LoadInt32(&p.waiting))
}

func (p *GPoolWithFunc) Cap() int {
	return int(atomic.LoadInt32(&p.cap))
}

// Running 当前正在运行的goroutine数量
func (p *GPoolWithFunc) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free 当前空闲的goroutine数量，-1表示无上限
func (p *GPoolWithFunc) Free() int {
	c := p.Cap()
	if c < 0 {
		return -1
	}
	return c - p.Running()
}

// Release 关闭pool
func (p *GPoolWithFunc) Release() {
	if !atomic.CompareAndSwapInt32(&p.state, 0, 1) {
		return
	}
	p.lock.Lock()
	p.workers.reset()
	p.lock.Unlock()

	p.cond.Broadcast()
}

// Reboot 重启一个pool
func (p *GPoolWithFunc) Reboot() {
	if atomic.CompareAndSwapInt32(&p.state, 1, 0) {
		// todo: 重启成功后的逻辑
	}
}

// Tune 更改池的容量
func (p *GPoolWithFunc) Tune(size int) {
	capacity := p.Cap()
	if capacity == -1 || size <= 0 || size == capacity {
		return
	}
	atomic.StoreInt32(&p.cap, int32(size))
	if size > capacity {
		if size-capacity == 1 {
			p.cond.Signal()
			return
		}
		p.cond.Broadcast()
	}
}

func (p *GPoolWithFunc) addRunning(delta int) {
	atomic.AddInt32(&p.running, int32(delta))
}

func (p *GPoolWithFunc) addWaiting(delta int) {
	atomic.AddInt32(&p.waiting, int32(delta))
}

// Invoke 唤醒一个任务
func (p *GPoolWithFunc) Invoke(args interface{}) error {
	if p.IsClosed() {
		return errors.New("pool is closed")
	}
	if w := p.retrieveWorker(); w != nil {
		w.inputParam(args)
		return nil
	}
	return errors.New("pool is overloaded")
}
