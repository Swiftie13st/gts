//go:build linux

/**
  @author: Bruce
  @since: 2023/12/6
  @desc: //TODO
**/

package epoll

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"gts/iface"
	"net"
	"sync"
	"syscall"
)

type ConnManager struct {
	connections map[uint64]iface.IConnection //管理的连接信息
	connLock    sync.RWMutex                 //读写连接的读写锁
	fd          int                          // epoll 描述符
	fd2conn     map[int]uint64               // fd 转 connID
	// epConnections map[int]net.Conn // 存储连接的映射
}

func NewConnManager() *ConnManager {
	fd, err := unix.EpollCreate1(0) // 创建 epoll 描述符
	if err != nil {
		fmt.Println("EpollCreate1 err: ", err)
		return nil
	}

	cm := &ConnManager{
		connections: make(map[uint64]iface.IConnection),
		fd2conn:     make(map[int]uint64),
		fd:          fd,
	}

	return cm
}

// Add 添加链接
func (cm *ConnManager) Add(conn iface.IConnection) {
	tcpConn, ok := conn.GetConnection().(*net.TCPConn)
	if !ok {
		fmt.Println("epoll conn is not net.Conn")
		return
	}
	fd := getFD(tcpConn)
	event := unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)}
	err := unix.EpollCtl(cm.fd, syscall.EPOLL_CTL_ADD, fd, &event)
	if err != nil {
		fmt.Println("EpollCtl Add err: ", err)
		return
	}
	cm.connLock.Lock() // 不能defer解锁，Len方法会加锁
	// 将conn加入集合
	cm.connections[conn.GetConnID()] = conn
	cm.fd2conn[fd] = conn.GetConnID()
	cm.connLock.Unlock()
	fmt.Println("connection add to ConnManager successfully: conn num = ", cm.Len())
}

// Remove 删除连接
func (cm *ConnManager) Remove(conn iface.IConnection) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	tcpConn, ok := conn.GetConnection().(*net.TCPConn)
	if !ok {
		fmt.Println("epoll conn is not net.Conn")
		return
	}
	// 将conn移除
	fd := getFD(tcpConn)
	err := unix.EpollCtl(cm.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		fmt.Println("EpollCtl Del err: ", err)
		return
	}
	delete(cm.connections, conn.GetConnID())
	delete(cm.fd2conn, fd)

	fmt.Println("connection Remove ConnID=", conn.GetConnID(), " successfully: conn num = ", cm.Len())
}

// Get 利用ConnID获取链接
func (cm *ConnManager) Get(connID uint64) (iface.IConnection, error) {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if conn, ok := cm.connections[connID]; ok {
		return conn, nil
	}

	return nil, errors.New("connection not found")
}

// Len 获取当前连接数量
func (cm *ConnManager) Len() int {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	return len(cm.connections)
}

// ClearConn 删除并停止所有链接
func (cm *ConnManager) ClearConn() {
	cm.connLock.Lock()
	//停止并删除全部的连接信息
	for _, conn := range cm.connections {
		//停止
		cm.Remove(conn)
	}
	cm.connLock.Unlock()
	fmt.Println("Clear All Connections successfully: conn num = ", cm.Len())
}

// GetAllConnID 获取所有连接ID
func (cm *ConnManager) GetAllConnID() []uint64 {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	ids := make([]uint64, 0, len(cm.connections))

	for id := range cm.connections {
		ids = append(ids, id)
	}

	return ids
}

// StartEpollWait 等待事件发生并返回相关的连接
func (cm *ConnManager) StartEpollWait() ([]iface.IConnection, error) {
	events := make([]unix.EpollEvent, 100)
retry:
	n, err := unix.EpollWait(cm.fd, events, 100)
	if err != nil {
		if err == unix.EINTR {
			goto retry
		}
		fmt.Println("EpollWait err: ", err)
		return nil, err
	}
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	var connections []iface.IConnection
	for i := 0; i < n; i++ {
		fd := int(events[i].Fd)
		conn := cm.connections[cm.fd2conn[fd]]
		connections = append(connections, conn)
	}
	return connections, nil
}
