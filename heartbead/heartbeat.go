/**
  @author: Bruce
  @since: 2023/5/18
  @desc: //TODO
**/

package heartbead

import (
	"errors"
	"fmt"
	"gts/iface"
	"gts/router"
	"time"
)

type Heartbeat struct {
	// 心跳检测时间间隔
	interval time.Duration
	// 退出信号
	quitChan chan struct{}
	// 心跳的消息ID
	msgID uint32
	//用户自定义的心跳检测消息业务处理路由
	router iface.IRouter
	//用户自定义的心跳检测消息处理方法
	makeMsg iface.HeartBeatMsgFunc
	//用户自定义的远程连接不存活时的处理方法
	onRemoteNotAlive iface.OnRemoteNotAlive
	// 绑定的链接
	conn iface.IConnection
}

// 默认的心跳消息生成函数
func makeDefaultMsg(conn iface.IConnection) []byte {
	fmt.Println("makeDefaultMsg")
	msg := fmt.Sprintf("heartbeat [%s->%s]", conn.LocalAddr(), conn.RemoteAddr())
	return []byte(msg)
}

// 默认的心跳检测函数
func notAliveDefaultFunc(conn iface.IConnection) {
	fmt.Printf("Remote connection %s is not alive, stop it", conn.RemoteAddr())
	conn.Stop()

}

// NewHeartbeat 创建心跳检测器
func NewHeartbeat(interval time.Duration) iface.IHeartbeat {
	fmt.Println("NewHeartbeat", interval)
	heartbeat := &Heartbeat{
		interval: interval,
		quitChan: make(chan struct{}),

		//均使用默认的心跳消息生成函数和远程连接不存活时的处理方法
		makeMsg:          makeDefaultMsg,
		onRemoteNotAlive: notAliveDefaultFunc,
		msgID:            iface.HeartBeatDefaultMsgID,
		router:           &router.HeatBeatDefaultRouter{},
	}

	return heartbeat
}

func (h *Heartbeat) SetOnRemoteNotAlive(f iface.OnRemoteNotAlive) {
	if f != nil {
		h.onRemoteNotAlive = f
	}
}

func (h *Heartbeat) SetHeartbeatMsgFunc(f iface.HeartBeatMsgFunc) {
	if f != nil {
		h.makeMsg = f
	}
}

func (h *Heartbeat) BindRouter(msgID uint32, router iface.IRouter) {
	if router != nil && msgID != iface.HeartBeatDefaultMsgID {
		h.msgID = msgID
		h.router = router
	}
}

// Start 启动心跳检测
func (h *Heartbeat) Start() {
	go func() {
		ticker := time.NewTicker(h.interval)
		fmt.Println("Start Heartbeat")
		for {
			select {
			case <-ticker.C:
				err := h.check()
				if err != nil {
					fmt.Println("Heartbeat err: ", err)
					return
				}
			case <-h.quitChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止心跳检测
func (h *Heartbeat) Stop() {
	fmt.Printf("heartbeat checker stop, connID=%+v", h.conn.GetConnID())
	h.quitChan <- struct{}{}
}

func (h *Heartbeat) SendHeartBeatMsg() error {

	msg := h.makeMsg(h.conn)
	fmt.Println("SendHeartBeatMsg: ", msg)
	err := h.conn.Send(h.msgID, msg)
	if err != nil {
		fmt.Printf("send heartbeat msg error: %v, msgId=%+v msg=%+v", err, h.msgID, msg)
		return err
	}

	return nil
}

// 执行心跳检测
func (h *Heartbeat) check() (err error) {

	if h.conn == nil {
		return errors.New("conn is not exist")
	}

	if !h.conn.IsAlive() {
		h.onRemoteNotAlive(h.conn)
		h.quitChan <- struct{}{}
	} else {
		err = h.SendHeartBeatMsg()
	}

	return err
}

// BindConn 绑定一个链接 TODO 多个链接？
func (h *Heartbeat) BindConn(conn iface.IConnection) {
	fmt.Println("BindConn: ", conn.GetConnID())
	h.conn = conn
	conn.SetHeartBeat(h)
}

func (h *Heartbeat) GetMsgID() uint32 {
	return h.msgID
}

func (h *Heartbeat) GetRouter() iface.IRouter {
	return h.router
}

// Clone 克隆到一个指定的链接上
func (h *Heartbeat) Clone() iface.IHeartbeat {
	heatbeat := &Heartbeat{
		interval:         h.interval,
		quitChan:         make(chan struct{}),
		makeMsg:          h.makeMsg,
		onRemoteNotAlive: h.onRemoteNotAlive,
		msgID:            h.msgID,
		router:           h.router,
		conn:             nil, //绑定的链接需要重新赋值
	}

	return heatbeat
}
