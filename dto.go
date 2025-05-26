package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
)

type CommandType string

const (
	CommandType_Subscribe   CommandType = "subscribe"
	CommandType_Unsubscribe CommandType = "unsubscribe"
)

type WebsocketCommand struct {
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WebsocketMessage struct {
	Commands []WebsocketCommand `json:"commands"`
}

type ClientManager struct {
	conns sync.Map
}

func (cm *ClientManager) Broadcast(msg WebsocketCommand) {
	cm.conns.Range(func(key, value interface{}) bool {
		_ = value.(*websocket.Conn).WriteMessage(websocket.TextMessage, msg.Payload)
		return true
	})
}
