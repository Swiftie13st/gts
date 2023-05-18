package main

import (
	"fmt"
	"gts/iface"
	"gts/server"
	"io"
	"net"
	"testing"
)

// ClientRouter  自定义路由
type ClientRouter struct {
	server.BaseRouter //一定要先基础BaseRouter
}

func (cr *ClientRouter) Handle(request iface.IRequest) {
	fmt.Println("Client Test ... start")

	//向服务器端写数据
	err := request.GetConnection().Send(1, []byte("Hello world"))
	if err != nil {
		return
	}
	err = request.GetConnection().Send(99999, []byte("Hello world2222"))
	if err != nil {
		return
	}
	go recvMsg(request.GetConnection().GetTCPConnection())
	select {}
}

/*
	模拟客户端
*/
func ClientTest() {

	fmt.Println("Client Test ... start")

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

	for i := 0; i < 2; i++ {
		//创建一个封包对象 dp
		dp := server.NewDataPack()

		msg, err := dp.Pack(server.NewMsgPackage(1, []byte("Hello world")))
		if err != nil {
			fmt.Println("Pack error msg id = ", 1)
			return
		}
		msg2, err := dp.Pack(server.NewMsgPackage(99999, []byte("Hello world2222")))
		if err != nil {
			fmt.Println("Pack error msg id = ", 2)
			return
		}
		fmt.Println("send: ", msg, msg2)
		//向服务器端写数据
		conn.Write(msg)
		conn.Write(msg2)

	}
	go recvMsg(conn)

	select {}
}

func recvMsg(conn net.Conn) {
	dp := server.NewDataPack()
	for {
		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read head error")
			break
		}
		//将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err:", err)
			return
		}
		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*server.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}

			fmt.Println("==> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}
	}

}

//Server 模块的测试函数
func TestServer(t *testing.T) {
	client := server.NewClient("127.0.0.1", 7777)

	client.AddRouter(1, &ClientRouter{})
	//启动心跳检测
	client.StartHeartBeat()
	client.Start()

	select {}
}
