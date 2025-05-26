package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var clientManager ClientManager

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func websocketCommandsHandler(commands []WebsocketCommand) error {
	for _, command := range commands {
		switch command.Type {
		case CommandType_Subscribe:
			var payloads []string
			if err := json.Unmarshal(command.Payload, &payloads); err != nil {
				fmt.Println("Error in unmarshalling payload of subscribing", err)
				return err
			}
			subscribeCommand(payloads)
		case CommandType_Unsubscribe:
			var payloads []string
			if err := json.Unmarshal(command.Payload, &payloads); err != nil {
				fmt.Println("Error in unmarshalling payload of unsubscribing", err)
				return err
			}
			unsubscribeCommand(payloads)
		default:
			return errors.New("Unknown Command Type")
		}
	}
	return nil
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {

	// Handshake
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error in http handshake", err)
		return
	}

	userID := rand.Uint64()
	clientManager.conns.Store(userID, conn)
	fmt.Println("Client connected", userID)

	var cleanupOnce sync.Once
	connectionClose := make(chan struct{})
	cleanupFn := func() {
		cleanupOnce.Do(func() {
			close(connectionClose)
			clientManager.conns.Delete(userID)
			if err := conn.Close(); err != nil {
				fmt.Println("Error in closing websocket connection")
			}
			fmt.Println("Connection closed", userID)
		})
	}

	// ping pong handler
	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Minute)); err != nil {
		fmt.Println("Error setting read deadline", err)
		return
	}

	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(1 * time.Minute)); err != nil {
			fmt.Println("Error setting read deadline", err)
			return err
		}
		return nil
	})

	go func() {
		ticker := time.NewTicker(time.Second * 25)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
					cleanupFn()
				}
			case <-connectionClose:
				return
			}
		}
	}()

	// Clean up resource
	defer func() {
		cleanupFn()
	}()

	// Receive Message
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				break
			}
			fmt.Println("Error in reading message", err)
			continue
		}
		if msgType == websocket.TextMessage {
			var data WebsocketMessage
			if err := json.Unmarshal(msg, &data); err != nil {
				fmt.Println("Error in unmarshalling message", err)
				continue
			}
			// Direct Message To Correct Command Handler
			if err := websocketCommandsHandler(data.Commands); err != nil {
				fmt.Println("Error in handling message", err)
				continue
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", websocketHandler)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
