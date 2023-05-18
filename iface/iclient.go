/**
  @author: Bruce
  @since: 2023/5/18
  @desc: //TODO
**/

package iface

// IClient 定义服务器接口
type IClient interface {
	// Start 启动服务器方法
	Start()
	// Stop 停止服务器方法
	Stop()
	// Conn 当前连接
	Conn() IConnection
	// AddRouter 路由功能：给当前服务注册一个路由业务方法
	AddRouter(msgid uint32, router IRouter)

	// GetMsgHandler 得到消息处理器
	GetMsgHandler() IMsgHandle

	// SetOnConnStart 设置该Client的连接创建时Hook函数
	SetOnConnStart(func(IConnection))
	// SetOnConnStop 设置该Client的连接断开时的Hook函数
	SetOnConnStop(func(IConnection))
	// GetOnConnStart 得到该Client的连接创建时Hook函数
	GetOnConnStart() func(IConnection)
	// GetOnConnStop 得到该Client的连接断开时的Hook函数
	GetOnConnStop() func(IConnection)

	// StartHeartBeat 启动心跳检测
	StartHeartBeat()
}
