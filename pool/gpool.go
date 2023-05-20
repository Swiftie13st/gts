/**
  @author: Bruce
  @since: 2023/5/21
  @desc: //TODO
**/

package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

//实现了对大规模 goroutine 的调度管理、goroutine 复用，
//允许使用者在开发并发程序的时候限制 goroutine 数量，
//复用资源，达到更高效执行任务的效果。

type Pool struct {
	// 上限容量
	cap int32
	// 当前正在运行的goroutine
	running int32

	lock sync.Locker
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
}

// 获取一个worker
func (p *Pool) retrieveWorker() (w worker) {
	spawnWorker := func() {
		w = p.workerCache.Get().(*goWorker)
		w.run()
	}

	p.lock.Lock()
	w = p.workers.detach()
	if w != nil { // first try to fetch the worker from the queue
		p.lock.Unlock()
	} else if capacity := p.Cap(); capacity == -1 || capacity > p.Running() {
		// 创建一个新worker
		p.lock.Unlock()
		spawnWorker()
	} else {
		// 已达上限，等待有worker放回
	retry:

		p.addWaiting(1)
		p.cond.Wait() // block and wait for an available worker
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
func (p *Pool) revertWorker(worker *goWorker) bool {
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

func NewPool(size int) (*Pool, error) {
	if size <= 0 {
		size = -1
	}

	p := &Pool{
		cap:  int32(size),
		lock: NewSpinLock(),
	}
	p.workerCache.New = func() interface{} {
		return &goWorker{
			pool: p,
			task: make(chan func(), 1),
		}
	}
	p.cond = sync.NewCond(p.lock)
	p.workers = newWorkerStack(size)
	return p, nil
}

func (p *Pool) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == 1
}

// Waiting 当前正在等待运行的goroutine数量
func (p *Pool) Waiting() int {
	return int(atomic.LoadInt32(&p.waiting))
}

func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.cap))
}

// Running 当前正在运行的goroutine数量
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free 当前空闲的goroutine数量，-1表示无上限
func (p *Pool) Free() int {
	c := p.Cap()
	if c < 0 {
		return -1
	}
	return c - p.Running()
}

// Release 关闭pool
func (p *Pool) Release() {
	if !atomic.CompareAndSwapInt32(&p.state, 0, 1) {
		return
	}
	p.lock.Lock()
	p.workers.reset()
	p.lock.Unlock()

	p.cond.Broadcast()
}

// Reboot 重启一个pool
func (p *Pool) Reboot() {
	if atomic.CompareAndSwapInt32(&p.state, 1, 0) {
		// todo: 重启成功后的逻辑
	}
}

// Tune 更改池的容量
func (p *Pool) Tune(size int) {
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

func (p *Pool) addRunning(delta int) {
	atomic.AddInt32(&p.running, int32(delta))
}

func (p *Pool) addWaiting(delta int) {
	atomic.AddInt32(&p.waiting, int32(delta))
}

// Submit 提交一个任务
func (p *Pool) Submit(task func()) error {
	if p.IsClosed() {
		return errors.New("pool is closed")
	}
	if w := p.retrieveWorker(); w != nil {
		w.inputFunc(task)
		return nil
	}
	return errors.New("pool is overloaded")
}
