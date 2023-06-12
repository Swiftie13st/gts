/**
  @author: Bruce
  @since: 2023/6/12
  @desc: //protobuf格式
**/

package server

// DataPackPB 封包拆包类实例，暂时不需要成员
type DataPackPB struct{}

// NewDataPackPB 封包拆包实例初始化方法
func NewDataPackPB() *DataPackPB {
	return &DataPackPB{}
}
