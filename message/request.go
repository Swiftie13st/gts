/**
  @author: Bruce
  @since: 2023/4/1
  @desc: //TODO
**/

package message

import "gts/iface"

type Request struct {
	Conn iface.IConnection //已经和客户端建立好的链接
	Msg  iface.IMessage    //客户端请求的数据
}

// GetConnection 获取请求连接信息
func (r *Request) GetConnection() iface.IConnection {
	return r.Conn
}

// GetData 获取请求消息的数据
func (r *Request) GetData() []byte {
	return r.Msg.GetData()
}

// GetMsgID 获取请求的消息的ID
func (r *Request) GetMsgID() uint32 {
	return r.Msg.GetMsgId()
}
