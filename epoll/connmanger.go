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
	"gts/iface"
	"net"
	"sync"
)

type ConnManager struct {
	connections map[uint64]iface.IConnection //管理的连接信息
	connLock    sync.RWMutex                 //读写连接的读写锁
	ep          *epoll                       // epoller
}

func NewConnManager() *ConnManager {
	ep, err := NewEpoll()
	if err != nil {
		panic(err)
	}
	return &ConnManager{
		connections: make(map[uint64]iface.IConnection),
		ep:          ep,
	}
}

// Add 添加链接
func (cm *ConnManager) Add(conn iface.IConnection) {
	cm.connLock.Lock()
	// 将conn加入集合
	cm.connections[conn.GetConnID()] = conn
	tcpConn, ok := conn.(net.Conn)
	if !ok {
		fmt.Println("epoll conn is not net.Conn")
	}
	err := cm.ep.Add(tcpConn)
	if err != nil {
		fmt.Println("epoll add err", err)
		err := tcpConn.Close()
		if err != nil {
			fmt.Println("conn close err", err)
			return
		}
	}

	cm.connLock.Unlock()
	fmt.Println("connection add to ConnManager successfully: conn num = ", cm.Len())
}

// Remove 删除连接
func (cm *ConnManager) Remove(conn iface.IConnection) {
	cm.connLock.Lock()

	// 将conn移除
	delete(cm.connections, conn.GetConnID())
	tcpConn, ok := conn.(net.Conn)
	if !ok {
		fmt.Println("epoll conn is not net.Conn")
	}
	err := cm.ep.Remove(tcpConn)
	if err != nil {
		fmt.Println("epoll remove conn err", err)
		return
	}

	cm.connLock.Unlock()
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
	for connID, conn := range cm.connections {
		//停止
		conn.Stop()
		delete(cm.connections, connID)
		tcpConn, ok := conn.(net.Conn)
		if !ok {
			fmt.Println("epoll conn is not net.Conn")
		}
		err := cm.ep.Remove(tcpConn)
		if err != nil {
			fmt.Println("epoll remove conn err", err)
			return
		}
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
