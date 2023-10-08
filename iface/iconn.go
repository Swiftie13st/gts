/**
  @author: Bruce
  @since: 2023/3/31
  @desc: //连接接口
**/

package iface

import (
	"net"
)

// IConnection 定义连接接口
type IConnection interface {
	// Start 启动连接，让当前连接开始工作
	Start()
	// Stop 停止连接，结束当前连接状态M
	Stop()
	// GetConnection 获取连接
	GetConnection() interface{}
	// GetConnID 获取当前连接ID
	GetConnID() uint64
	// RemoteAddr 获取远程客户端地址信息
	RemoteAddr() net.Addr
	// LocalAddr 获取链接本地地址信息
	LocalAddr() net.Addr
	// Send 直接将数据发送数据给远程的TCP客户端
	Send(msgId uint32, data []byte) error

	// SetProperty 设置链接属性
	SetProperty(key string, value interface{})
	// GetProperty 获取链接属性
	GetProperty(key string) (interface{}, error)
	// RemoveProperty 移除链接属性
	RemoveProperty(key string)
	// IsAlive 判断当前连接是否存活
	IsAlive() bool
	// SetHeartBeat 设置心跳检测器
	SetHeartBeat(checker IHeartbeat)
}
