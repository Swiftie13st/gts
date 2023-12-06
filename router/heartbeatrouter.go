/**
  @author: Bruce
  @since: 2023/5/18
  @desc: //TODO
**/

package router

import (
	"fmt"
	"gts/iface"
)

// HeatBeatDefaultRouter 收到remote心跳消息的默认回调路由业务
type HeatBeatDefaultRouter struct {
	BaseRouter
}

func (r *HeatBeatDefaultRouter) Handle(req iface.IRequest) {
	fmt.Printf("Recv Heartbeat from %s, MsgID = %+v, Data = %s",
		req.GetConnection().RemoteAddr(), req.GetMsgID(), string(req.GetData()))
}
