/*
*

	@author: Bruce
	@since: 2023/5/30
	@desc: //TODO

*
*/
package server

import (
	"fmt"
	"gts/iface"
	"gts/utils"
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
)

// PingRouter ping test 自定义路由
type PingRouter struct {
	BaseRouter //一定要先基础BaseRouter
}

func (pr *PingRouter) Handle(request iface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	err := request.GetConnection().Send(1, []byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error", err)
	}
}

func TestNewServer(t *testing.T) {
	go func() {
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	utils.InitSettings("../conf/config.yaml")
	s := NewServer()

	s.AddRouter(1, &PingRouter{})
	s.Serve()
}

func TestNewGServer(t *testing.T) {
	utils.InitSettings("../conf/config.yaml")
	s := NewGServer()

	s.AddRouter(1, &PingRouter{})
	s.Serve()
}
