package server

import (
	"fmt"
	"gts/iface"
	"net"
	"testing"
	"time"
)

/*
	模拟客户端
*/
func ClientTest() {

	fmt.Println("Client Test ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("client Close err, exit!")
		}
	}(conn)
	for i := 0; i < 5; i++ {
		_, err := conn.Write([]byte("hahaha"))
		if err != nil {
			fmt.Println("write error err ", err)
			return
		}

		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read buf error ")
			return
		}

		fmt.Printf(" server call back : %s, cnt = %d\n", buf[:cnt], cnt)

		time.Sleep(1 * time.Second)
	}
}

//ping test 自定义路由
type PingRouter struct {
	BaseRouter //一定要先基础BaseRouter
}

//Test PreHandle
func (pr *PingRouter) PreHandle(request iface.IRequest) {
	fmt.Println("Call Router PreHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping ....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Test Handle
func (pr *PingRouter) Handle(request iface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Test PostHandle
func (pr *PingRouter) PostHandle(request iface.IRequest) {
	fmt.Println("Call Router PostHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("After ping .....\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Server 模块的测试函数
func TestServer(t *testing.T) {

	/*
		服务端测试
	*/
	//1 创建一个server 句柄 s
	s := NewServer()
	s.AddRouter(&PingRouter{})
	/*
		客户端测试
	*/
	go ClientTest()
	//2 开启服务
	s.Serve()

}
