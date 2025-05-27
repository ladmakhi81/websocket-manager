package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var clientManager = NewClientManager()

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
			return errors.New("unknown command")
		}
	}
	return nil
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	userIDParam := r.Header.Get("user-id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Handshake
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("error in http handshake", err)
		return
	}

	clientManager.AddConnection(uint(userID), conn)
	fmt.Println("client connected", userID)

	var cleanupOnce sync.Once
	connectionClose := make(chan struct{})
	cleanupFn := func() {
		cleanupOnce.Do(func() {
			close(connectionClose)
			clientManager.RemoveConnection(uint(userID), conn)
			if err := conn.Close(); err != nil {
				fmt.Println("error in closing websocket connection")
			}
			fmt.Println("connection closed", userID)
		})
	}

	// ping pong handler
	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Minute)); err != nil {
		fmt.Println("error setting read deadline", err)
		return
	}

	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(1 * time.Minute)); err != nil {
			fmt.Println("error setting read deadline", err)
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
			fmt.Println("error in reading message", err)
			continue
		}
		if msgType == websocket.TextMessage {
			var data WebsocketMessage
			if err := json.Unmarshal(msg, &data); err != nil {
				fmt.Println("error in unmarshalling message", err)
				continue
			}
			// Direct Message To Correct Command Handler
			if err := websocketCommandsHandler(data.Commands); err != nil {
				fmt.Println("error in handling message", err)
				continue
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", websocketHandler)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
