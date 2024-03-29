//go:build linux

/**
  @author: Bruce
  @since: 2023/12/1
  @desc: //test epoll
**/

package epoll

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

func TestEpoll(t *testing.T) {
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}

	ep, err := NewEpoll()
	if err != nil {
		panic(err)
	}
	go func() {
		buf := make([]byte, 1024)
		for {
			connections, err := ep.Wait()
			if err != nil {
				continue
			}
			for _, conn := range connections {
				n, err := conn.Read(buf)
				if err != nil {
					if err := ep.Remove(conn); err != nil {
						break
					}
					err := conn.Close()
					if err != nil {
						break
					}
				}
				log.Println("recv:", n, string(buf))
			}
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			t.Log("accept err", err)
			return
		}

		if err := ep.Add(conn); err != nil {
			t.Log("epoll add err", err)
			err := conn.Close()
			if err != nil {
				t.Log("conn close err", err)
				return
			}
		}
	}
}

func TestEpollClient(t *testing.T) {
	addr := "127.0.0.1:8888"

	c, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		fmt.Println("failed to connect", err)
		return
	}

	defer func() {
		c.Close()
	}()

	for {
		time.Sleep(time.Second)
		//log.Printf("连接 %d 发送数据", i)
		c.Write([]byte("hello world\r\n"))
	}
}
