package cmd

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type WebSocketServer struct {
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	broadcast chan Message
	mu        sync.Mutex
	rdb       *redis.Client
}

// const (
// 	pingPeriod = 30 * time.Second
// 	pongWait   = 60 * time.Second
// 	writeWait  = 10 * time.Second
// 	maxMsgSize = 512
// )

func NewWebSocketServer(rdb *redis.Client) *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan Message),
		rdb:       rdb,
	}
}

func (wsServer *WebSocketServer) HandleConnections(c *gin.Context) {
	conn, err := wsServer.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("⚠️ WebSocket panic recovered:", r)
		}
		wsServer.mu.Lock()
		delete(wsServer.clients, conn)
		wsServer.mu.Unlock()
		conn.Close()
		log.Println("⚠️ Client disconnected")
	}()

	wsServer.mu.Lock()
	wsServer.clients[conn] = true
	wsServer.mu.Unlock()

	log.Println("✅ Client connected")

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("⚠️ Read error:", err)
			continue
		}
		// wsServer.broadcast <- msg
		// response := Message{
		// 	Username: "server",
		// 	Message:  "Message received",
		// }
		// time.Sleep(time.Second)
		// conn.WriteJSON(response)
		jsonMsg, _ := json.Marshal(msg)
		wsServer.rdb.Publish(context.Background(), "chatroom", jsonMsg)

		// Optional server confirmation
		ack := Message{Username: "server", Message: "✅ Message received"}
		time.Sleep(time.Second)
		conn.WriteJSON(ack)

	}
}

func (wsServer *WebSocketServer) HandleMessages() {
	for msg := range wsServer.broadcast {
		wsServer.BroadcastMessage(msg)

	}
}

func (wsServer *WebSocketServer) BroadcastMessage(msg Message) {
	wsServer.mu.Lock()
	defer wsServer.mu.Unlock()

	for client := range wsServer.clients {
		if err := client.WriteJSON(msg); err != nil {
			log.Println("Write Error:", err)
			client.Close()
			delete(wsServer.clients, client)
		}
	}
}

// func getAIResponse(userInput string) (string, error) {
// 	resp, err := openAIClient.CreateChatCompletion(
// 		context.Background(),
// 		openai.ChatCompletionRequest{
// 			Model: openai.GPT3Dot5Turbo,
// 			Messages: []openai.ChatCompletionMessage{
// 				{Role: "system", Content: "You are a helpful AI assistant."},
// 				{Role: "user", Content: userInput},
// 			},
// 		},
// 	)
// 	if err != nil {
// 		return "", err
// 	}
// 	return resp.Choices[0].Message.Content, nil
// }

// func runChannel() {
// 	randomStrings := []string{
// 		"default string 1",
// 		"default string 2",
// 		"default string 3",
// 		"default string 4",
// 	}
// 	rs := make(chan string)
// 	go func() {
// 		for v, i := range randomStrings {
// 			rs <- i
// 			time.Sleep(time.Second)
// 			log.Printf("send %d", v)
// 			getIndex(v)
// 		}
// 	}()
// }

// func getIndex(idx int) {
// 	idxChannel := make(chan int)
// 	go func() {
// 		idxChannel <- idx
// 		time.Sleep(time.Second)
// 	}()
// }
