package main

import (
	"chatbox/cmd"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	fmt.Print("Hello World")
	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	wsServer := cmd.NewWebSocketServer()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/chat", wsServer.HandleConnections)

	// Run message handling in a goroutine
	go wsServer.HandleMessages()
	r.Run(":4000")
}
