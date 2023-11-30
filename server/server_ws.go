/**
  @author: Bruce
  @since: 2023/11/30
  @desc: //TODO
**/

package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"gts/utils"
	"net/http"
)

func (s *Server) startWebSocketServer() {
	fmt.Printf("[START] WS Server listener at IP: %s, Port %d, is starting\n", s.IP, s.WsPort)

	http.HandleFunc("/", s.upgradeWs)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.IP, s.WsPort), nil)
	if err != nil {
		fmt.Println("start ws err: ", err)
		return
	}
}

func (s *Server) upgradeWs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("upgradeWs")
	if len(r.Header.Get("Sec-Websocket-Protocol")) > 0 {
		// todo Token校验
	}

	if s.ConnMgr.Len() >= utils.Conf.MaxConn {
		fmt.Println("Exceeded the maxConn")
		return
	}

	// 将HTTP连接升级为WebSocket连接
	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		//token 校验
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Subprotocols: []string{r.Header.Get("Sec-WebSocket-Protocol")},
	}).Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade WS err: ", err)
		return
	}
	cid, err := s.sf.NextVal()
	if err != nil {
		fmt.Println("Id gen err ", err)
		return
	}
	dealConn := newServerWsConn(s, conn, cid)
	go dealConn.Start()
}
