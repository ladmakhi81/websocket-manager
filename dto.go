package main

import (
	jsonEncoding "encoding/json"
)

type CommandType string

const (
	CommandType_Subscribe   CommandType = "subscribe"
	CommandType_Unsubscribe CommandType = "unsubscribe"
)

type WebsocketCommand struct {
	Type    CommandType             `json:"type"`
	Payload jsonEncoding.RawMessage `json:"payload"`
}

type WebsocketMessage struct {
	Commands []WebsocketCommand `json:"commands"`
}
