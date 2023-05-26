/**
  @author: Bruce
  @since: 2023/5/26
  @desc: //TODO 限制并发连接数，定时回收旧连接
**/

package iface

type IRateLimit interface {
	// SetOnLimitReached 设置当达到限制时触发的Hook函数
	SetOnLimitReached()
}

// OnLimitReached 用户自定义的远程连接达到限制时的处理方法
type OnLimitReached func(IConnection)
