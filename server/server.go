/**
  @author: Bruce
  @since: 2023/3/17
  @desc: //服务器核心
**/

package server

import (
	"context"
	"fmt"
	"gts/iface"
	"gts/utils"
	"os"
	"os/signal"
)

// Server 接口实现，定义一个Server服务类
type Server struct {
	//服务器的名称
	Name string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port   int
	WsPort int
	//当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	msgHandler iface.IMsgHandle
	//当前Server的链接管理器
	ConnMgr iface.IConnManager

	//该Server的连接创建时Hook函数
	onConnStart func(conn iface.IConnection)
	//该Server的连接断开时的Hook函数
	onConnStop func(conn iface.IConnection)

	//心跳检测器
	hb iface.IHeartbeat

	sf *utils.SnowflakeGenerator
	// 捕获链接关闭状态
	exitChan chan struct{}
}

// NewServer 创建一个服务器句柄
func NewServer() iface.IServer {
	s := &Server{
		Name:       utils.Conf.Name,
		IPVersion:  utils.Conf.IpVersion,
		IP:         utils.Conf.Ip,
		Port:       utils.Conf.Port,
		WsPort:     utils.Conf.WsPort,
		msgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
		sf:         utils.NewSnowflakeGenerator(utils.Conf.WorkerId, utils.Conf.DatacenterId),
	}

	return s
}

//============== 实现 iface.IServer 里的全部接口方法 ========

func (s *Server) Start() {

	s.exitChan = make(chan struct{})

	//0 启动worker工作池机制
	s.msgHandler.StartWorkerPool()
	if utils.Conf.TCPMode {
		go s.startTcpServer()
	}
	if utils.Conf.WSMode {
		go s.startWebSocketServer()
	}
	if utils.Conf.QuicMode {
		go s.startQuicServer(context.Background())
	}

}

func (s *Server) Stop() {
	fmt.Println("[STOP] Gts server , name ", s.Name)
	s.ConnMgr.ClearConn()
	// 直接close 让所有exitChan都读到空值
	close(s.exitChan)
}

func (s *Server) Serve() {
	s.Start()

	// 阻塞,否则主Go退出， listener的go将会退出
	c := make(chan os.Signal, 1)
	// 监听指定信号 ctrl+c kill信号
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	fmt.Printf("[SERVE] %s, Serve Interrupt, signal = %v", s.Name, sig)
}

// AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(msgId uint32, router iface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)

	fmt.Println("Add Router success! ")
}

func (s *Server) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}
func (s *Server) GetMsgHandler() iface.IMsgHandle {
	return s.msgHandler
}

// SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(iface.IConnection)) {
	s.onConnStart = hookFunc
}

// SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(iface.IConnection)) {
	s.onConnStop = hookFunc
}

// GetOnConnStart 得到该Server的连接创建时Hook函数
func (s *Server) GetOnConnStart() func(iface.IConnection) {
	return s.onConnStart
}

// GetOnConnStop 得到该Server的连接断开时的Hook函数
func (s *Server) GetOnConnStop() func(iface.IConnection) {
	return s.onConnStop
}
func (s *Server) GetHeartBeat() iface.IHeartbeat {
	return s.hb
}

// StartHeartBeat 启动心跳检测
func (s *Server) StartHeartBeat() {
	hb := NewHeartbeat(utils.Conf.GetHeartbeatInterval())
	s.AddRouter(hb.GetMsgID(), hb.GetRouter())
	s.hb = hb
}
