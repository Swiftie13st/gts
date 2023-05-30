/**
  @author: Bruce
  @since: 2023/4/1
  @desc: //服务器接口
**/

package iface

// IServer 定义服务器接口
type IServer interface {
	// Start 启动服务器方法
	Start()
	// Stop 停止服务器方法
	Stop()
	// Serve 开启业务服务方法
	Serve()
	// AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
	AddRouter(uint32, IRouter)
	// GetConnMgr 得到链接管理
	GetConnMgr() IConnManager
	// GetMsgHandler 得到消息处理器
	GetMsgHandler() IMsgHandle

	// SetOnConnStart 设置该Server的连接创建时Hook函数
	SetOnConnStart(func(IConnection))
	// SetOnConnStop 设置该Server的连接断开时的Hook函数
	SetOnConnStop(func(IConnection))
	// GetOnConnStart 得到该Server的连接创建时Hook函数
	GetOnConnStart() func(IConnection)
	// GetOnConnStop 得到该Server的连接断开时的Hook函数
	GetOnConnStop() func(IConnection)

	// StartHeartBeat 启动心跳检测
	StartHeartBeat()
	// GetHeartBeat 获取心跳检测器
	GetHeartBeat() IHeartbeat
}
