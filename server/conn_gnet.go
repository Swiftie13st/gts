/**
  @author: Bruce
  @since: 2023/5/31
  @desc: //TODO
**/

package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/gnet"
	"gts/iface"
	"gts/message"
	"gts/utils"
	"net"
	"sync"
	"time"
)

type GConnection struct {

	//当前连接的socket TCP套接字
	Conn gnet.Conn

	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint64
	//当前连接的关闭状态
	isClosed bool

	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler iface.IMsgHandle

	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc

	//用于读、写两个goroutine之间的消息通信
	msgChan chan []byte
	//用户收发消息的Lock
	msgLock sync.RWMutex

	//当前conn属于那个ConnManger
	connManager iface.IConnManager
	//当前连接创建时Hook函数
	onConnStart func(conn iface.IConnection)
	//当前连接断开时的Hook函数
	onConnStop func(conn iface.IConnection)

	//链接属性
	property map[string]interface{}
	//保护当前property的锁
	propertyLock sync.Mutex

	//心跳检测器
	hb iface.IHeartbeat
	//最后一次活动时间
	lastActivityTime time.Time
}

// newGServerConn 创建连接的方法
func newGServerConn(server iface.IServer, conn gnet.Conn, connID uint64) iface.IConnection {
	c := &GConnection{
		Conn:     conn,
		ConnID:   connID,
		isClosed: false,
		//ExitBuffChan: make(chan bool, 1),
		MsgHandler:  server.GetMsgHandler(),
		msgChan:     make(chan []byte, 1024),
		connManager: server.GetConnMgr(),
		onConnStart: server.GetOnConnStart(),
		onConnStop:  server.GetOnConnStop(),
	}

	server.GetConnMgr().Add(c)
	return c
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *GConnection) StartWriter() {
	fmt.Println("Writer Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " Conn Writer exit!")
	defer c.Stop()

	for {
		select {
		case data := <-c.msgChan:
			fmt.Println("StartWriter msgChan: ", data)
			err := c.Conn.SendTo(data)
			if err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
			//有数据要写给客户端
			//if _, err := c.Conn.Write(data); err != nil {
			//	fmt.Println("Send Data error:, ", err, " Conn Writer exit")
			//	return
			//}
		case <-c.ctx.Done():
			fmt.Println("StartWriter ExitBuffChan")
			//conn已经关闭
			return
		}
	}
}

// StartReader 读消息Goroutine，用于从客户端中读取数据
func (c *GConnection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " Conn reader exit!")
	defer c.Stop()

	// TODO
	return
	for {
		// 创建拆包解包的对象
		dp := message.NewDataPack()

		//读取客户端的Msg head

		conn := *c.GetGConnection()
		size, headData := conn.ReadN(int(dp.GetHeadLen()))
		if size != int(dp.GetHeadLen()) {
			fmt.Println("read Msg head length err, length : ", size)
			//c.ExitBuffChan <- true
			return
		}

		fmt.Println("headData", string(headData))

		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.UnpackHead(headData)
		if err != nil {
			fmt.Println("unpack err ", err)
			//c.ExitBuffChan <- true
			continue
		}

		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			size, data = conn.ReadN(int(msg.GetDataLen()))
			if size != int(msg.GetDataLen()) {
				fmt.Println("read Msg data length err, length : ", size)
				//c.ExitBuffChan <- true
				return
			}
		}
		msg.SetData(data)
		if msg.GetMsgId() == iface.HeartBeatDefaultMsgID {
			//心跳检测数据，更新心跳检测Active状态
			fmt.Println("心跳检测数据，更新心跳检测Active状态")
			if c.hb != nil {
				c.updateActivity()
			}
		} else {
			//得到当前客户端请求的Request数据
			fmt.Println("得到当前客户端请求的Request数据")
			req := message.Request{
				Conn: c,
				Msg:  msg,
			}
			if utils.Conf.WorkerPoolSize > 0 {
				//已经启动工作池机制，将消息交给Worker处理
				c.MsgHandler.SendMsgToTaskQueue(&req)
			} else {
				//从绑定好的消息和对应的处理方法中执行对应的Handle方法
				go c.MsgHandler.DoMsgHandler(&req)
			}
		}
	}

}

// Start 启动连接，让当前连接开始工作
func (c *GConnection) Start() {
	fmt.Println("Conn Start(), ConnID = ", c.ConnID)

	//启动心跳检测
	if c.hb != nil {
		c.hb.Start()
		c.updateActivity()
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	//1 开启用于写回客户端数据流程的Goroutine
	go c.StartReader()
	//2 开启用户从客户端读取数据流程的Goroutine
	go c.StartWriter()
	c.callOnConnStart()
	for {
		select {
		case <-c.ctx.Done():

			//得到退出消息，不再阻塞
			return
		}
	}
}

// Stop 停止连接，结束当前连接状态M
func (c *GConnection) Stop() {
	fmt.Println("Conn Stop(), ConnID = ", c.ConnID)
	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	//关闭链接绑定的心跳检测器
	if c.hb != nil {
		c.hb.Stop()
	}
	c.isClosed = true
	c.callOnConnStop()
	if c.connManager != nil {
		c.connManager.Remove(c)
	}

	// 关闭socket链接
	err := c.Conn.Close()
	if err != nil {
		fmt.Println("Conn Stop() Error, err = ", err)
		return
	}

	//通知从缓冲队列读数据的业务，该链接已经关闭
	c.cancel()
	//close(c.ExitBuffChan)
	//close(c.msgChan)
}

func (c *GConnection) GetConnection() interface{} {
	return c.Conn
}

// GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *GConnection) GetTCPConnection() *net.TCPConn {
	return nil
}

func (c *GConnection) GetWSConnection() *websocket.Conn {
	return nil
}

func (c *GConnection) GetGConnection() *gnet.Conn {
	return &c.Conn
}

// GetConnID 获取当前连接ID
func (c *GConnection) GetConnID() uint64 {
	return c.ConnID
}

// RemoteAddr 获取远程客户端地址信息
func (c *GConnection) RemoteAddr() net.Addr {

	return c.Conn.RemoteAddr()
}

// LocalAddr 获取链接本地地址信息
func (c *GConnection) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

// Send 直接将数据封包发送数据给远程的TCP客户端
func (c *GConnection) Send(msgId uint32, data []byte) error {
	c.msgLock.RLock()
	defer c.msgLock.RUnlock()

	fmt.Println("Send ", string(data))
	if c.isClosed == true {
		return errors.New("connection closed when send Msg")
	}
	//将data封包，并且发送
	dp := message.NewDataPack()
	msg, err := dp.Pack(message.NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error Msg id = ", msgId)
		return errors.New("Pack error Msg ")
	}

	//写回客户端
	c.msgChan <- msg

	return nil
}

// callOnConnStart 调用连接OnConnStart Hook函数
func (c *GConnection) callOnConnStart() {
	if c.onConnStart != nil {
		fmt.Println("CallOnConnStart....")
		c.onConnStart(c)
	}
}

// callOnConnStart 调用连接OnConnStop Hook函数
func (c *GConnection) callOnConnStop() {
	if c.onConnStop != nil {
		fmt.Println("CallOnConnStop....")
		c.onConnStop(c)
	}
}

// SetProperty 设置链接属性
func (c *GConnection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	if c.property == nil {
		c.property = make(map[string]interface{})
	}

	c.property[key] = value
}

// GetProperty 获取链接属性
func (c *GConnection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	}

	return nil, errors.New("no property found")
}

// RemoveProperty 移除链接属性
func (c *GConnection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

func (c *GConnection) SetHeartBeat(hb iface.IHeartbeat) {
	c.hb = hb
}

func (c *GConnection) IsAlive() bool {
	if c.isClosed {
		return false
	}
	lastTimeInterval := time.Now().Sub(c.lastActivityTime)
	fmt.Println("isAlive: ", lastTimeInterval)
	// 检查连接最后一次活动时间，如果超过心跳间隔，则认为连接已经死亡
	return lastTimeInterval < utils.Conf.GetHeartbeatMaxTime()
}

func (c *GConnection) updateActivity() {
	c.lastActivityTime = time.Now()
}
