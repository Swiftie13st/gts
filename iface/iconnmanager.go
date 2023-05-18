/**
  @author: Bruce
  @since: 2023/5/18
  @desc: //链接管理接口
**/

package iface

type IConnManager interface {
	Add(conn IConnection)                   //添加链接
	Remove(conn IConnection)                //删除连接
	Get(connID uint64) (IConnection, error) //利用ConnID获取链接
	Len() int                               //获取当前连接数量
	ClearConn()                             //删除并停止所有链接
	GetAllConnID() []uint64                 //获取所有连接ID
}
