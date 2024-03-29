/**
  @author: Bruce
  @since: 2023/5/18
  @desc: //管理连接信息
**/

package server

import (
	"errors"
	"fmt"
	"gts/iface"
	"sync"
)

type ConnManager struct {
	connections map[uint64]iface.IConnection //管理的连接信息
	connLock    sync.RWMutex                 //读写连接的读写锁
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint64]iface.IConnection),
	}
}

// Add 添加链接
func (cm *ConnManager) Add(conn iface.IConnection) {
	cm.connLock.Lock()
	// 将conn加入集合
	cm.connections[conn.GetConnID()] = conn
	cm.connLock.Unlock()
	fmt.Println("connection add to ConnManager successfully: Conn num = ", cm.Len())
}

// Remove 删除连接
func (cm *ConnManager) Remove(conn iface.IConnection) {
	cm.connLock.Lock()

	// 将conn移除
	delete(cm.connections, conn.GetConnID())
	cm.connLock.Unlock()
	fmt.Println("connection Remove ConnID=", conn.GetConnID(), " successfully: Conn num = ", cm.Len())
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
	}
	cm.connLock.Unlock()
	fmt.Println("Clear All Connections successfully: Conn num = ", cm.Len())
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
