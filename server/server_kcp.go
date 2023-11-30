/**
  @author: Bruce
  @since: 2023/11/30
  @desc: //KCP
**/

package server

import (
	"crypto/sha1"
	"fmt"
	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"gts/utils"
)

func (s *Server) startKCPServer() {
	fmt.Printf("[START] Quic Server listener at IP: %s, Port %d, is starting\n", s.IP, s.Port)
	addr := fmt.Sprintf("%s:%d", s.IP, s.Port)

	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	listener, err := kcp.ListenWithOptions(addr, block, 10, 3)
	if err != nil {
		fmt.Println("listen kcp", addr, "err", err)
		return
	}
	//已经监听成功
	fmt.Println("start Gts KCP server  ", s.Name, " success, now listening...")
	//3 启动server网络连接业务
	go func() {
		for {
			//服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.Len() >= utils.Conf.MaxConn {
				fmt.Println("Exceeded the maxConn")
				continue
			}
			//阻塞等待客户端建立连接请求
			conn, err := listener.AcceptKCP()
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
			dealConn := newKCPServerConn(s, conn, cid)

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
