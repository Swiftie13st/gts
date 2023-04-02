/**
  @author: Bruce
  @since: 2023/4/1
  @desc: //
**/

package iface

/*
	路由接口， 这里面路由是 使用框架者给该链接自定的 处理业务方法
	路由里的IRequest 则包含用该链接的链接信息和该链接的请求数据信息
*/
type IRouter interface {
	// PreHandle 在处理conn业务之前的钩子方法Hook
	PreHandle(request IRequest)
	
	// Handle 处理conn业务的方法
	Handle(request IRequest)

	// PostHandle 处理conn业务之后的钩子方法
	PostHandle(request IRequest)
}
