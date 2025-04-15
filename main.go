package main

import (
	"chatbox/cmd"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	wsServer := cmd.NewWebSocketServer(rdb)
	// wsServer := cmd.NewWebSocketServer()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/chat", wsServer.HandleConnections)

	// Run message handling in a goroutine
	go wsServer.HandleMessages()
	go func() {
		pubsub := rdb.Subscribe(context.Background(), "chatroom")
		ch := pubsub.Channel()
		for msg := range ch {
			var incoming cmd.Message
			if err := json.Unmarshal([]byte(msg.Payload), &incoming); err != nil {
				log.Println(" Redis Unmarshal Error:", err)
				continue
			}
			wsServer.BroadcastMessage(incoming)
		}
	}()
	r.Run(":4000")
}
