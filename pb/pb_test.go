/**
  @author: Bruce
  @since: 2023/6/12
  @desc: //TODO
**/

package msg

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"gts/pb/msg"
	"testing"
)

var msg1 = &msg.Message{
	ID:      1,
	DataLen: 8,
	Data:    "123123",
}

var data []byte = []byte{8, 1, 16, 8, 26, 6, 49, 50, 51, 49, 50, 51}

func TestPB_Marshal(t *testing.T) {
	data, _ = proto.Marshal(msg1)

	fmt.Println(data, string(data))
}

func TestPB_Unmarshal(t *testing.T) {
	msg2 := &msg.Message{}

	err := proto.Unmarshal(data, msg2)

	fmt.Println(msg2, err)
}
