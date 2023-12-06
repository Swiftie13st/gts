//go:build linux
// +build linux

/**
  @author: Bruce
  @since: 2023/12/1
  @desc: Implements a TCP server using epoll on Linux.
**/

package epoll

import (
	"fmt"
	"net"
	"reflect"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type epoll struct {
	fd          int              // epoll 描述符
	connections map[int]net.Conn // 存储连接的映射
	lock        *sync.RWMutex    // 用于保护 connections 的读写锁
}

// NewEpoll 创建一个新的 epoll 实例
func NewEpoll() (*epoll, error) {
	fd, err := unix.EpollCreate1(0) // 创建 epoll 描述符
	if err != nil {
		fmt.Println("EpollCreate1 err: ", err)
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
	}, nil
}

// Add 将连接添加到 epoll 实例中
func (e *epoll) Add(conn net.Conn) error {
	fd := getFD(conn)
	event := unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)}
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &event)
	if err != nil {
		fmt.Println("EpollCtl Add err: ", err)
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn

	fmt.Println("添加连接，连接总数：", len(e.connections))

	return nil
}

// Remove 从 epoll 实例中移除连接
func (e *epoll) Remove(conn net.Conn) error {
	fd := getFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		fmt.Println("EpollCtl Del err: ", err)
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, fd)

	fmt.Println("删除连接，连接总数：", len(e.connections))

	return nil
}

// Wait 等待事件发生并返回相关的连接
func (e *epoll) Wait() ([]net.Conn, error) {
	events := make([]unix.EpollEvent, 100)
retry:
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		if err == unix.EINTR {
			goto retry
		}
		fmt.Println("EpollWait err: ", err)
		return nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Fd)]
		connections = append(connections, conn)
	}
	return connections, nil
}

// getFD 获取连接的文件描述符
func getFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")

	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}
