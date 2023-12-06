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
	"gts/message"
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

	dp := message.NewDataPack()
	msg, err := dp.Pack(message.NewMsgPackage(1, []byte("Hello world")))
	if err != nil {
		fmt.Println("Pack error Msg id = ", 1)
		return
	}
	err = c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		fmt.Println("write Msg err: ", err)
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
		p := message.NewDataPack()
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
	defer func(stream quic.Stream) {
		err := stream.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(stream)
	data := []byte("Hello, server!")

	db := message.NewDataPack()
	pack, err := db.Pack(message.NewMsgPackage(1, data))
	if err != nil {
		fmt.Println("pack err", err)
		return
	}

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := stream.Read(buf)
			if err != nil {
				fmt.Println("quic io read error", err, n)
				return
			}

			fmt.Printf("Received message: %s\n", string(buf[:n]))
		}

	}()

	for i := 0; i < 4; i++ {
		n, err := stream.Write(pack)
		if err != nil {
			fmt.Println("send pack err", err)
			return
		}
		fmt.Println("send pack", pack, n)
		time.Sleep(time.Second * 2)

	}
	time.Sleep(time.Second * 10)
}
