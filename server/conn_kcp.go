/**
  @author: Bruce
  @since: 2023/11/30
  @desc: //kcp Conn
**/

package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/xtaci/kcp-go"
	"gts/iface"
	"gts/message"
	"gts/utils"
	"net"
	"sync"
	"time"
)

type ConnectionKCP struct {

	//当前连接
	Conn *kcp.UDPSession
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
func newKCPServerConn(server iface.IServer, conn *kcp.UDPSession, connID uint64) iface.IConnection {
	c := &ConnectionKCP{
		Conn:     conn,
		ConnID:   connID,
		isClosed: false,
		//ExitBuffChan: make(chan bool, 1),
		MsgHandler:  server.GetMsgHandler(),
		msgChan:     make(chan []byte),
		connManager: server.GetConnMgr(),
		onConnStart: server.GetOnConnStart(),
		onConnStop:  server.GetOnConnStop(),
	}

	server.GetConnMgr().Add(c)
	return c
}

// newClientConn 创建连接的方法
func newKCPClientConn(client iface.IClient, conn *kcp.UDPSession) iface.IConnection {
	c := &ConnectionKCP{
		Conn:     conn,
		ConnID:   0,
		isClosed: false,
		//ExitBuffChan: make(chan bool, 1),
		MsgHandler:  client.GetMsgHandler(),
		msgChan:     make(chan []byte),
		onConnStart: client.GetOnConnStart(),
		onConnStop:  client.GetOnConnStop(),
	}

	return c
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *ConnectionKCP) StartWriter() {
	fmt.Println("Writer Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " Conn Writer exit!")
	defer c.Stop()

	for {
		select {
		case data := <-c.msgChan:
			fmt.Println("StartWriter msgChan: ", data)
			_, err := c.Conn.Write(data)
			if err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case <-c.ctx.Done():
			fmt.Println("StartWriter ExitBuffChan")
			//conn已经关闭
			return
		}
	}
}

// StartReader 读消息Goroutine，用于从客户端中读取数据
func (c *ConnectionKCP) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " Conn reader exit!")
	defer c.Stop()
	// 创建拆包解包的对象
	dp := message.NewDataPack()

	conn, ok := c.GetConnection().(*kcp.UDPSession)
	if !ok {
		fmt.Println("get kcp Conn err", c.GetConnection())
		//c.ExitBuffChan <- true
		return
	}
	//(*Conn).SendMessage()

	for {
		headData := make([]byte, dp.GetHeadLen())
		size, err := conn.Read(headData)
		if size != int(dp.GetHeadLen()) {
			fmt.Println("read Msg head length err, length : ", size)
			//c.ExitBuffChan <- true
			return
		}
		//fmt.Println("headData", headData, dp.GetHeadLen())
		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.UnpackHead(headData)
		if err != nil {
			fmt.Println("unpack err ", err)
			//c.ExitBuffChan <- true
			continue
		}

		//根据 dataLen 读取 data，放在msg.Data中
		data := make([]byte, msg.GetDataLen())
		if msg.GetDataLen() > 0 {
			size, err = conn.Read(data)
			if err != nil || size != int(msg.GetDataLen()) {
				fmt.Println("read Msg data length err, length : , err: ", size, err)
				//c.ExitBuffChan <- true
				return
			}
		}
		msg.SetData(data)
		fmt.Println("data", string(data))
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
func (c *ConnectionKCP) Start() {
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

	select {
	case <-c.ctx.Done():

		//得到退出消息，不再阻塞
		return
	}

}

// Stop 停止连接，结束当前连接状态M
func (c *ConnectionKCP) Stop() {
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

func (c *ConnectionKCP) GetConnection() interface{} {
	return c.Conn
}

// GetConnID 获取当前连接ID
func (c *ConnectionKCP) GetConnID() uint64 {
	return c.ConnID
}

// RemoteAddr 获取远程客户端地址信息
func (c *ConnectionKCP) RemoteAddr() net.Addr {

	return c.Conn.RemoteAddr()
}

// LocalAddr 获取链接本地地址信息
func (c *ConnectionKCP) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

// Send 直接将数据封包发送数据给远程的TCP客户端
func (c *ConnectionKCP) Send(msgId uint32, data []byte) error {
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
func (c *ConnectionKCP) callOnConnStart() {
	if c.onConnStart != nil {
		fmt.Println("CallOnConnStart....")
		c.onConnStart(c)
	}
}

// callOnConnStart 调用连接OnConnStop Hook函数
func (c *ConnectionKCP) callOnConnStop() {
	if c.onConnStop != nil {
		fmt.Println("CallOnConnStop....")
		c.onConnStop(c)
	}
}

// SetProperty 设置链接属性
func (c *ConnectionKCP) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	if c.property == nil {
		c.property = make(map[string]interface{})
	}

	c.property[key] = value
}

// GetProperty 获取链接属性
func (c *ConnectionKCP) GetProperty(key string) (interface{}, error) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	}

	return nil, errors.New("no property found")
}

// RemoveProperty 移除链接属性
func (c *ConnectionKCP) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

func (c *ConnectionKCP) SetHeartBeat(hb iface.IHeartbeat) {
	c.hb = hb
}

func (c *ConnectionKCP) IsAlive() bool {
	if c.isClosed {
		return false
	}
	lastTimeInterval := time.Now().Sub(c.lastActivityTime)
	fmt.Println("isAlive: ", lastTimeInterval)
	// 检查连接最后一次活动时间，如果超过心跳间隔，则认为连接已经死亡
	return lastTimeInterval < utils.Conf.GetHeartbeatMaxTime()
}

func (c *ConnectionKCP) updateActivity() {
	c.lastActivityTime = time.Now()
}
