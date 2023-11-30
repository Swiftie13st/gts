/**
  @author: Bruce
  @since: 2023/11/30
  @desc: //TCP
**/

package server

import (
	"fmt"
	"gts/utils"
	"net"
)

func (s *Server) startTcpServer() {
	fmt.Printf("[START] Tcp Server listener at IP: %s, Port %d, is starting\n", s.IP, s.Port)
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
	go func() {
		for {
			//服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.Len() >= utils.Conf.MaxConn {
				fmt.Println("Exceeded the maxConn")
				continue
			}
			//阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}

			//3.3 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
			cid, err := s.sf.NextVal()
			if err != nil {
				fmt.Println("Id gen err ", err)
				continue
			}
			dealConn := newServerConn(s, conn, cid)

			//HeartBeat 心跳检测
			if s.hb != nil {
				//从Server端克隆一个心跳检测器
				heartBeat := s.hb.Clone()
				//绑定当前链接
				heartBeat.BindConn(dealConn)
			}

			//3.4 启动当前链接的处理业务
			go dealConn.Start()
		}
	}()
	select {
	case <-s.exitChan:
		err := listener.Close()
		if err != nil {
			fmt.Println("listener close err: ", err)
		}
	}
}
