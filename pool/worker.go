/**
  @author: Bruce
  @since: 2023/5/21
  @desc: //worker
**/

package pool

import (
	"fmt"
	"runtime/debug"
	"time"
)

type goWorker struct {
	// 属于的协程池
	pool *GPool
	// 用于接收异步任务包的chan
	task chan func()

	lastUsed time.Time
}

func (w *goWorker) run() {
	w.pool.addRunning(1)
	go func() {
		defer func() {
			w.pool.addRunning(-1)
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				fmt.Printf("worker exits from panic: %v\n%s\n", p, debug.Stack())
			}
			// 唤醒等待队列
			w.pool.cond.Signal()
		}()

		for f := range w.task {
			if f == nil {
				return
			}
			f()
			if ok := w.pool.revertWorker(w); !ok {
				//fmt.Println("")
				return
			}
		}
	}()
}

func (w *goWorker) finish() {
	w.task <- nil
}

func (w *goWorker) lastUsedTime() time.Time {
	return w.lastUsed
}

func (w *goWorker) inputFunc(fn func()) {
	w.task <- fn
}
func (w *goWorker) inputParam(interface{}) {
	panic("unreachable")
}
