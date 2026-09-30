package realtime

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"task-management-api/pkg/response"
	"task-management-api/pkg/token"
)

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
	secret  string
	upgrade websocket.Upgrader
}

func NewHub(secret string) *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]struct{}),
		secret:  secret,
		upgrade: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

func (h *Hub) Handle(c *gin.Context) {
	value := c.Query("token")
	if _, err := token.Parse(value, h.secret); value == "" || err != nil {
		response.Error(c, http.StatusUnauthorized, "valid token query parameter required")
		return
	}
	connection, err := h.upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	h.clients[connection] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, connection)
		h.mu.Unlock()
		_ = connection.Close()
	}()

	for {
		messageType, message, readErr := connection.ReadMessage()
		if readErr != nil {
			return
		}
		h.Broadcast(messageType, message)
	}
}

func (h *Hub) Broadcast(messageType int, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		if err := client.WriteMessage(messageType, message); err != nil {
			_ = client.Close()
			delete(h.clients, client)
		}
	}
}
