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

// PingRouter ping test 自定义路由
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

// Ping2Router ping test 自定义路由
type Ping2Router struct {
	server.BaseRouter //一定要先基础BaseRouter
}

func (pr *Ping2Router) Handle(request iface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	err := request.GetConnection().Send(1, []byte("ping2...ping2...ping2\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

func handleStart(conn iface.IConnection) {
	fmt.Println("Start")
	err := conn.Send(202, []byte("handleStart"))
	if err != nil {
		return
	}
}

func handleStop(conn iface.IConnection) {
	fmt.Println("Stop")
	err := conn.Send(202, []byte("handleStop"))
	if err != nil {
		return
	}
}
func main() {
	utils.InitSettings("./conf/config.yaml")

	s := server.NewServer()

	s.SetOnConnStart(handleStart)
	s.SetOnConnStop(handleStop)

	s.AddRouter(1, &PingRouter{})
	s.AddRouter(2, &Ping2Router{})
	s.Serve()
}
