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
	AddRouter(msgid uint32, router IRouter)
	// GetConnMgr 得到链接管理
	GetConnMgr() IConnManager
	// GetMsgHandler 得到消息处理器
	GetMsgHandler() IMsgHandle
}
