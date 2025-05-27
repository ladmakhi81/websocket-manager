package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type ClientManager struct {
	conns sync.Map
	mu    sync.Mutex
}

func (cm *ClientManager) AddConnection(userID uint, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	connections, ok := cm.conns.Load(userID)
	if !ok {
		cm.conns.Store(userID, []*websocket.Conn{conn})
		return
	}
	newConnections := append(connections.([]*websocket.Conn), conn)
	cm.conns.Store(userID, newConnections)
}

func (cm *ClientManager) RemoveConnection(userID uint, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	connections, ok := cm.conns.Load(userID)
	if !ok {
		return
	}
	updatedConnections := connections.([]*websocket.Conn)
	for index, storedConn := range updatedConnections {
		if storedConn == conn {
			updatedConnections[index] = updatedConnections[len(updatedConnections)-1]
			updatedConnections = updatedConnections[:len(updatedConnections)-1]
		}
	}
	cm.conns.Store(userID, updatedConnections)
}

func (cm *ClientManager) Broadcast(msg WebsocketCommand) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conns.Range(func(key, values any) bool {
		connections := values.([]*websocket.Conn)
		for _, conn := range connections {
			if err := conn.WriteMessage(websocket.TextMessage, msg.Payload); err != nil {
				fmt.Println("error writing message", err)
			}
		}
		return true
	})
}
