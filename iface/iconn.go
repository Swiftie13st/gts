/**
  @author: Bruce
  @since: 2023/3/31
  @desc: //TODO
**/

package iface

import "net"

// IConnection 定义连接接口
type IConnection interface {
	// Start 启动连接，让当前连接开始工作
	Start()
	// Stop 停止连接，结束当前连接状态M
	Stop()
	// GetTCPConnection 从当前连接获取原始的socket TCPConn
	GetTCPConnection() *net.TCPConn
	// GetConnID 获取当前连接ID
	GetConnID() uint32
	// RemoteAddr 获取远程客户端地址信息
	RemoteAddr() net.Addr
	// Send 直接将数据发送数据给远程的TCP客户端
	Send(msgId uint32, data []byte) error
}

// HandFunc 定义一个统一处理链接业务的接口 连接 数据 长度
type HandFunc func(*net.TCPConn, []byte, int) error
