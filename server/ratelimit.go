/**
  @author: Bruce
  @since: 2023/5/26
  @desc: //TODO 限制并发连接数，定时回收旧连接
**/

package server

import (
	"fmt"
	"gts/iface"
)

type RateLimit struct {
}

func NewRateLimit(MaxConn int64) iface.IRateLimit {
	fmt.Println("NewHeartbeat", MaxConn)

	return &RateLimit{}
}

func (r *RateLimit) SetOnLimitReached() {

}
