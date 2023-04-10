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

func (pr *PingRouter) PreHandle(request iface.IRequest) {
	fmt.Println("Call Router PreHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping ....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

func (pr *PingRouter) Handle(request iface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

func (pr *PingRouter) PostHandle(request iface.IRequest) {
	fmt.Println("Call Router PostHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("After ping .....\n"))
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
