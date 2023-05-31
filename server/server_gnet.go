/**
  @author: Bruce
  @since: 2023/5/30
  @desc: //测试gnet
**/

package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
	"gts/iface"
	"gts/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
)

// GServer 接口实现，定义一个Server服务类
type GServer struct {
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
	// gnet
	*gnet.EventServer

	multicore bool
	pool      *goroutine.Pool
}

// NewGServer 创建一个服务器句柄
func NewGServer() iface.IServer {
	s := &GServer{
		Name:       utils.Conf.Name,
		IPVersion:  utils.Conf.IpVersion,
		IP:         utils.Conf.Ip,
		Port:       utils.Conf.Port,
		WsPort:     utils.Conf.WsPort,
		msgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
		sf:         utils.NewSnowflakeGenerator(utils.Conf.WorkerId, utils.Conf.DatacenterId),
		multicore:  true,
	}

	return s
}

func (s *GServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Printf("Echo server is listening on %s (multi-cores: %t, loops: %d)\n",
		srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

// React 每次收到信息后操作
func (s *GServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(1, frame))
	if err != nil {
		fmt.Println("data pack err:", err)
		return
	}
	fmt.Println(string(msg))
	out = msg

	//_ = s.pool.Submit(func() {
	//	time.Sleep(1 * time.Second)
	//	err := c.AsyncWrite(msg)
	//	if err != nil {
	//		fmt.Println("React AsyncWrite err: ", err)
	//		return
	//	}
	//})

	return
}

// OnOpened fires when a new connection has been opened.
// The parameter out is the return value which is going to be sent back to the peer.
func (s *GServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	//服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
	if s.ConnMgr.Len() >= utils.Conf.MaxConn {
		fmt.Println("Exceeded the maxConn")
		return
	}
	cid, err := s.sf.NextVal()
	if err != nil {
		fmt.Println("Id gen err ", err)
		return
	}
	dealConn := newGServerConn(s, c, cid)

	//HeartBeat 心跳检测
	if s.hb != nil {
		//从Server端克隆一个心跳检测器
		heartBeat := s.hb.Clone()
		//绑定当前链接
		heartBeat.BindConn(dealConn)
	}

	//3.4 启动当前链接的处理业务
	go dealConn.Start()

	return
}

func (s *GServer) startTcpServer() {
	fmt.Printf("[START] gnet Tcp Server listener at IP: %s, Port %d, is starting\n", s.IP, s.Port)
	p := goroutine.Default()
	defer p.Release()
	s.pool = p

	err := gnet.Serve(s, fmt.Sprintf("tcp://:%d", s.Port), gnet.WithMulticore(s.multicore))
	if err != nil {
		fmt.Println("gnet err: ", err)
		return
	}

	select {
	case <-s.exitChan:
		if err != nil {
			fmt.Println("listener close err: ", err)
		}
	}
}

func (s *GServer) startWebSocketServer() {
	fmt.Printf("[START] gnet WS Server listener at IP: %s, Port %d, is starting\n", s.IP, s.WsPort)

	http.HandleFunc("/", s.upgradeWs)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.IP, s.WsPort), nil)
	if err != nil {
		panic(err)
	}
}

func (s *GServer) upgradeWs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upgradeWs")
	if len(r.Header.Get("Sec-Websocket-Protocol")) > 0 {
		// todo Token校验
	}

	if s.ConnMgr.Len() >= utils.Conf.MaxConn {
		fmt.Println("Exceeded the maxConn")
		return
	}

	// 将HTTP连接升级为WebSocket连接
	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		//token 校验
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Subprotocols: []string{r.Header.Get("Sec-WebSocket-Protocol")},
	}).Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade WS err: ", err)
		return
	}
	cid, err := s.sf.NextVal()
	if err != nil {
		fmt.Println("Id gen err ", err)
		return
	}
	dealConn := newServerWsConn(s, conn, cid)
	go dealConn.Start()
}

//============== 实现 iface.IServer 里的全部接口方法 ========

func (s *GServer) Start() {

	s.exitChan = make(chan struct{})

	//0 启动worker工作池机制
	s.msgHandler.StartWorkerPool()
	if utils.Conf.TCPMode {
		go s.startTcpServer()
	}
	if utils.Conf.WSMode {
		//go s.startWebSocketServer()
	}

}

func (s *GServer) Stop() {
	fmt.Println("[STOP] Gts server , name ", s.Name)
	s.ConnMgr.ClearConn()
	s.exitChan <- struct{}{}
	close(s.exitChan)
}

func (s *GServer) Serve() {
	s.Start()

	// 阻塞,否则主Go退出， listener的go将会退出
	c := make(chan os.Signal, 1)
	// 监听指定信号 ctrl+c kill信号
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	fmt.Printf("[SERVE] %s, Serve Interrupt, signal = %v", s.Name, sig)
}

// AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *GServer) AddRouter(msgId uint32, router iface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)

	fmt.Println("Add Router success! ")
}

func (s *GServer) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}
func (s *GServer) GetMsgHandler() iface.IMsgHandle {
	return s.msgHandler
}

// SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *GServer) SetOnConnStart(hookFunc func(iface.IConnection)) {
	s.onConnStart = hookFunc
}

// SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *GServer) SetOnConnStop(hookFunc func(iface.IConnection)) {
	s.onConnStop = hookFunc
}

// GetOnConnStart 得到该Server的连接创建时Hook函数
func (s *GServer) GetOnConnStart() func(iface.IConnection) {
	return s.onConnStart
}

// GetOnConnStop 得到该Server的连接断开时的Hook函数
func (s *GServer) GetOnConnStop() func(iface.IConnection) {
	return s.onConnStop
}
func (s *GServer) GetHeartBeat() iface.IHeartbeat {
	return s.hb
}

// StartHeartBeat 启动心跳检测
func (s *GServer) StartHeartBeat() {
	hb := NewHeartbeat(utils.Conf.GetHeartbeatInterval())
	s.AddRouter(hb.GetMsgID(), hb.GetRouter())
	s.hb = hb
}
