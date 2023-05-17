/**
  @author: Bruce
  @since: 2023/3/17
  @desc: //服务器核心
**/

package server

import (
	"fmt"
	"gts/iface"
	"gts/utils"
	"net"
	"time"
)

//Server 接口实现，定义一个Server服务类
type Server struct {
	//服务器的名称
	Name string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	msgHandler iface.IMsgHandle
}

//============== 实现 iface.IServer 里的全部接口方法 ========

func (s *Server) Start() {
	fmt.Printf("[START] Server listener at IP: %s, Port %d, is starting\n", s.IP, s.Port)

	//开启一个go去做服务端Listener业务
	go func() {
		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err: ", err)
			return
		}

		//2 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		//已经监听成功
		fmt.Println("start Gts server  ", s.Name, " success, now listening...")

		//3 启动server网络连接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}

			//3.2 TODO Server.Start() 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接

			//3.3 TODO Server.Start() 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的

			//TODO server.go 应该有一个自动生成ID的方法
			var cid uint32
			cid = 0

			//3.3 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
			dealConn := NewConnection(conn, cid, s.msgHandler)
			cid++

			//3.4 启动当前链接的处理业务
			go dealConn.Start()
			time.Sleep(time.Second * 5)

		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Gts server , name ", s.Name)

	//TODO  Server.Stop() 将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
}

func (s *Server) Serve() {
	s.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listener的go将会退出
	select {}
}

// AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(msgId uint32, router iface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)

	fmt.Println("Add Router success! ")
}

// NewServer 创建一个服务器句柄
func NewServer() iface.IServer {
	s := &Server{
		Name:       utils.Conf.Name,
		IPVersion:  utils.Conf.IpVersion,
		IP:         utils.Conf.Ip,
		Port:       utils.Conf.Port,
		msgHandler: NewMsgHandle(),
	}

	return s
}

//func init() {
//	utils.InitSettings("../conf/config.yaml")
//}
