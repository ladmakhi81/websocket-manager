package main

import (
	"encoding/json"
	"fmt"
)

func subscribeCommand(markets []string) {
	clientManager.Broadcast(WebsocketCommand{Type: "somethong", Payload: json.RawMessage(`{'name': "nima"}`)})
	fmt.Println("Subscribing to Markets", markets)
}

func unsubscribeCommand(markets []string) {
	fmt.Println("UnSubscribing to Markets", markets)
}
