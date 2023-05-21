/**
  @author: Bruce
  @since: 2023/5/21
  @desc: //TODO
**/

package pool

import (
	"fmt"
	"runtime/debug"
	"time"
)

type goWorkerWithFunc struct {
	// 属于的协程池
	pool *GPoolWithFunc
	// 用于接收异步任务包的chan
	args chan interface{}

	lastUsed time.Time
}

func (w *goWorkerWithFunc) run() {
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

		for args := range w.args {
			if args == nil {
				return
			}
			w.pool.poolFunc(args)
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
	}()
}

func (w *goWorkerWithFunc) finish() {
	w.args <- nil
}

func (w *goWorkerWithFunc) lastUsedTime() time.Time {
	return w.lastUsed
}

func (w *goWorkerWithFunc) inputFunc(fn func()) {
	panic("unreachable")
}
func (w *goWorkerWithFunc) inputParam(arg interface{}) {
	w.args <- arg
}
