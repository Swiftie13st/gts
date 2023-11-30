/**
  @author: Bruce
  @since: 2023/11/30
  @desc: //Quic
**/

package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/quic-go/quic-go"
	"gts/utils"
	"math/big"
)

func (s *Server) startQuicServer(ctx context.Context) {
	fmt.Printf("[START] Quic Server listener at IP: %s, Port %d, is starting\n", s.IP, s.Port)
	addr := fmt.Sprintf("%s:%d", s.IP, s.Port)

	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		fmt.Println("listen quic", addr, "err", err)
		return
	}

	//已经监听成功
	fmt.Println("start Gts quic server  ", s.Name, " success, now listening...")
	//3 启动server网络连接业务
	go func() {
		for {
			//服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.Len() >= utils.Conf.MaxConn {
				fmt.Println("Exceeded the maxConn")
				continue
			}
			//阻塞等待客户端建立连接请求
			conn, err := listener.Accept(ctx)
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
			dealConn := newQuicServerConn(s, conn, cid)

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

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}
