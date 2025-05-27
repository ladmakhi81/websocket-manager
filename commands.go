package main

import (
	jsonEncoding "encoding/json"
	"fmt"
)

func subscribeCommand(markets []string) {
	clientManager.Broadcast(WebsocketCommand{Type: "somethong", Payload: jsonEncoding.RawMessage(`{'name': "nima"}`)})
	fmt.Println("Subscribing to Markets", markets)
}

func unsubscribeCommand(markets []string) {
	fmt.Println("UnSubscribing to Markets", markets)
}
