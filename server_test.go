package main

import (
	"fmt"
	"gts/server"
	"io"
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
		//创建一个封包对象 dp
		dp := server.NewDataPack()

		msg, err := dp.Pack(server.NewMsgPackage(1, []byte("Hello world")))
		if err != nil {
			fmt.Println("Pack error msg id = ", 1)
			return
		}

		fmt.Println("send: ", msg)
		//向服务器端写数据
		conn.Write(msg)
		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
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
	ClientTest()
}
