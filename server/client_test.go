/**
  @author: Bruce
  @since: 2023/5/30
  @desc: //TODO
**/

package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"testing"
)

func TestClient_WS(t *testing.T) {
	u := url.URL{Scheme: "ws", Host: "localhost:8888", Path: "/"}
	fmt.Println("ws connecting to: ", u)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"User-Agent": {""}})
	if err != nil {
		fmt.Println("dial:", err)
		return
	}

	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(1, []byte("Hello world")))
	if err != nil {
		fmt.Println("Pack error msg id = ", 1)
		return
	}
	err = c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		fmt.Println("write msg err: ", err)
		return
	}
	go read(c)
	select {}

}
func read(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		p := NewDataPack()
		img, err := p.Unpack(message)
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		fmt.Printf("msgId:%d,recv: %s", img.GetMsgId(), img.GetData())
	}
}
