/**
  @author: Bruce
  @since: 2023/5/18
  @desc: //TODO
**/

package server

import (
	"fmt"
	"gts/heartbead"
	"gts/iface"
	"gts/message"
	"gts/utils"
	"net"
	"time"
)

// Client 接口实现，定义一个Client服务类
type Client struct {
	//tcp4 or other
	IPVersion string
	//Client绑定的IP地址
	IP string
	//Client绑定的端口
	Port int
	//客户端链接
	conn iface.IConnection
	//当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	msgHandler iface.IMsgHandle

	//该Server的连接创建时Hook函数
	onConnStart func(conn iface.IConnection)
	//该Server的连接断开时的Hook函数
	onConnStop func(conn iface.IConnection)

	//心跳检测器
	hb iface.IHeartbeat
}

// NewClient 创建一个客户端句柄
func NewClient(ip string, port int) iface.IClient {
	c := &Client{
		IPVersion:  utils.Conf.IpVersion,
		IP:         ip,
		Port:       port,
		msgHandler: message.NewMsgHandle(),
	}

	return c
}

//============== 实现 iface.IServer 里的全部接口方法 ========

func (c *Client) Start() {

	// 客户端不需要workerPool
	utils.Conf.WorkerPoolSize = 0
	go func() {
		addr := &net.TCPAddr{
			IP:   net.ParseIP(c.IP),
			Port: c.Port,
			Zone: "", //for ipv6, ignore
		}
		//创建原始Socket，得到net.Conn
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			//创建链接失败
			fmt.Printf("client connect to server failed, err:%v\n", err)
			panic(err)
		}
		//创建Connection对象
		c.conn = newClientConn(c, conn)
		fmt.Printf("[START] Client at IP: %v, Port %d, is starting\n", c.IP, c.Port)
		//HeartBeat心跳检测
		if c.hb != nil {
			//创建链接成功，绑定链接与心跳检测器
			c.hb.BindConn(c.conn)
		}

		//启动链接
		go c.conn.Start()
		select {}

	}()
}

func (c *Client) Stop() {
	fmt.Printf("[STOP] Client Client LocalAddr: %c, RemoteAddr: %c\n", c.conn.LocalAddr(), c.conn.RemoteAddr())

	c.conn.Stop()
}

func (c *Client) Conn() iface.IConnection {
	return c.conn
}

// AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (c *Client) AddRouter(msgId uint32, router iface.IRouter) {
	c.msgHandler.AddRouter(msgId, router)

	fmt.Println("Add Router success! ")
}

func (c *Client) GetMsgHandler() iface.IMsgHandle {
	return c.msgHandler
}

// SetOnConnStart 设置该Server的连接创建时Hook函数
func (c *Client) SetOnConnStart(hookFunc func(iface.IConnection)) {
	c.onConnStart = hookFunc
}

// SetOnConnStop 设置该Server的连接断开时的Hook函数
func (c *Client) SetOnConnStop(hookFunc func(iface.IConnection)) {
	c.onConnStop = hookFunc
}

// GetOnConnStart 得到该Server的连接创建时Hook函数
func (c *Client) GetOnConnStart() func(iface.IConnection) {
	return c.onConnStart
}

// GetOnConnStop 得到该Server的连接断开时的Hook函数
func (c *Client) GetOnConnStop() func(iface.IConnection) {
	return c.onConnStop
}
func (c *Client) GetHeartBeat() iface.IHeartbeat {
	return c.hb
}

// StartHeartBeat 启动心跳检测
func (c *Client) StartHeartBeat() {
	hb := heartbead.NewHeartbeat(15 * time.Second)
	c.AddRouter(hb.GetMsgID(), hb.GetRouter())
	c.hb = hb
}
