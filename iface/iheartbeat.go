/**
  @author: Bruce
  @since: 2023/5/18
  @desc: //心跳监测机制
**/

package iface

const (
	HeartBeatDefaultMsgID uint32 = 99999
)

type IHeartbeat interface {
	Start()
	Stop()
	// Clone 复制一个一样的hb
	Clone() IHeartbeat
	SendHeartBeatMsg() error
	// SetOnRemoteNotAlive 心跳不活跃时
	SetOnRemoteNotAlive(OnRemoteNotAlive)
	// SetHeartbeatMsgFunc 心跳信息
	SetHeartbeatMsgFunc(HeartBeatMsgFunc)
	// BindRouter 绑定心跳路由
	BindRouter(uint32, IRouter)
	// BindConn 绑定一个链接
	BindConn(IConnection)
	// GetMsgID 获取心跳消息id
	GetMsgID() uint32
	// GetRouter 获取心跳路由
	GetRouter() IRouter
}

// HeartBeatMsgFunc 用户自定义的心跳检测消息处理方法
type HeartBeatMsgFunc func(IConnection) []byte

// OnRemoteNotAlive 用户自定义的远程连接不存活时的处理方法
type OnRemoteNotAlive func(IConnection)

type HeartBeatOption struct {
	MakeMsg          HeartBeatMsgFunc //用户自定义的心跳检测消息处理方法
	OnRemoteNotAlive OnRemoteNotAlive //用户自定义的远程连接不存活时的处理方法
	HeadBeatMsgID    uint32           //用户自定义的心跳检测消息ID
	Router           IRouter          //用户自定义的心跳检测消息业务处理路由
}
