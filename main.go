/**
  @author: Bruce
  @since: 2023/3/17
  @desc:
**/

package main

import (
	"fmt"
	"gts/iface"
	"gts/server"
	"gts/utils"
)

//ping test 自定义路由
type PingRouter struct {
	server.BaseRouter //一定要先基础BaseRouter
}

func (pr *PingRouter) Handle(request iface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	err := request.GetConnection().Send(1, []byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

func main() {
	utils.InitSettings("./conf/config.yaml")

	s := server.NewServer()
	s.AddRouter(&PingRouter{})
	s.Serve()
}
