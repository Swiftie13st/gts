/**
  @author: Bruce
  @since: 2023/5/17
  @desc: //消息管理模块
**/

package message

import (
	"fmt"
	"gts/iface"
	"gts/pool"
	"gts/utils"
	"strconv"
)

type MsgHandle struct {
	Apis           map[uint32]iface.IRouter //存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize int                      //业务工作Worker池的数量
	TaskQueue      []chan iface.IRequest    //Worker负责取任务的消息队列
	workerPool     *pool.GPoolWithFunc
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]iface.IRouter),
		WorkerPoolSize: utils.Conf.WorkerPoolSize,
		//一个worker对应一个queue
		TaskQueue: make([]chan iface.IRequest, utils.Conf.WorkerPoolSize),
	}
}

// DoMsgHandler 马上以非阻塞方式处理消息
func (mh *MsgHandle) DoMsgHandler(request iface.IRequest) {
	fmt.Println("DoMsgHandler")
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is not FOUND!")
		return
	}

	//执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// AddRouter 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router iface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	//2 添加msg与api的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}

// SendMsgToTaskQueue 将消息交给TaskQueue,由worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request iface.IRequest) {
	// 直接从池中取一个worker发送就好
	fmt.Println("SendMsgToTaskQueue Invoke")
	err := mh.workerPool.Invoke(request)
	if err != nil {
		fmt.Println("Invoke err: ", err)
		return
	}
}

// StartOneWorker 启动一个Worker工作流程
func (mh *MsgHandle) startOneWorker(workerID int, taskQueue chan iface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started.")
	//不断地等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// StartWorkerPool 启动worker工作池
func (mh *MsgHandle) StartWorkerPool() {
	mh.workerPool = pool.NewPoolWithFunc(mh.WorkerPoolSize, func(i interface{}) {
		mh.DoMsgHandler(i.(iface.IRequest))
	})
}
