package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type ClientManager struct {
	conns map[uint][]*websocket.Conn
	mu    sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{conns: make(map[uint][]*websocket.Conn)}
}

func (cm *ClientManager) AddConnection(userID uint, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	connections, ok := cm.conns[userID]
	if !ok {
		cm.conns[userID] = []*websocket.Conn{conn}
		return
	}
	connections = append(connections, conn)
	cm.conns[userID] = connections
}

func (cm *ClientManager) RemoveConnection(userID uint, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	connections, ok := cm.conns[userID]
	if !ok {
		return
	}
	for index, storedConn := range connections {
		if storedConn == conn {
			connections[index] = connections[len(connections)-1]
			connections = connections[:len(connections)-1]
		}
	}
	cm.conns[userID] = connections
}

func (cm *ClientManager) Broadcast(msg WebsocketCommand) {
	cm.mu.Lock()
	connections := make(map[uint][]*websocket.Conn, len(cm.conns))
	for userID, userConnections := range cm.conns {
		connections[userID] = userConnections
	}
	cm.mu.Unlock()
	for _, userIDConns := range connections {
		for _, conn := range userIDConns {
			if err := conn.WriteMessage(websocket.TextMessage, msg.Payload); err != nil {
				fmt.Println("error writing message", err)
			}
		}
	}
}
