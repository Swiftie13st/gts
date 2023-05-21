/**
  @author: Bruce
  @since: 2023/5/21
  @desc: //worker接口
**/

package pool

import "time"

type worker interface {
	run()
	finish()
	lastUsedTime() time.Time
	inputFunc(func())
	inputParam(interface{})
}

type workerQueue interface {
	// 当前数量
	len() int
	// 是否为空
	isEmpty() bool
	// 插入一个worker
	insert(worker) error
	// 获取一个worker
	detach() worker
	// 回收等待时间长的worker
	refresh(duration time.Duration) []worker
	// 重置
	reset()
}
