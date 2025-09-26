package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan gin.H)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) StatsWS(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	clients[ws] = true

	for {
		var msg gin.H
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}
	}
}

func broadcastUpdate(event gin.H) {
	for client := range clients {
		err := client.WriteJSON(event)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}