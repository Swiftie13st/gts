/**
  @author: Bruce
  @since: 2023/5/30
  @desc: //TODO
**/

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/quic-go/quic-go"
	"net/http"
	"net/url"
	"testing"
	"time"
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

func TestClient_Quic(t *testing.T) {
	ctx := context.Background()
	addr := "localhost:7777"
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	session, err := quic.DialAddr(ctx, addr, tlsConf, nil)
	if err != nil {
		fmt.Println("quic dial addr error", err)
		return
	}
	stream, err := session.OpenStream()
	if err != nil {
		fmt.Println("quic open stream error", err)
		return
	}
	defer stream.Close()
	data := []byte("Hello, server!")

	db := NewDataPack()
	pack, err := db.Pack(NewMsgPackage(1, data))
	if err != nil {
		fmt.Println("pack err", err)
		return
	}

	for i := 0; i < 4; i++ {
		n, err := stream.Write(pack)
		if err != nil {
			fmt.Println("send pack err", err)
			return
		}
		fmt.Println("send pack", pack, n)
		time.Sleep(time.Second * 2)

		buf := make([]byte, 1024)
		n, err = stream.Read(buf)
		if err != nil {
			fmt.Println("quic io read error", err, n)
			return
		}

		fmt.Printf("Received message: %s\n", string(buf[:n]))
	}

}
