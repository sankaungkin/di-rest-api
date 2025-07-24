package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sankangkin/di-rest-api/internal/domain/productstock"
)

type Client struct {
	Conn     *websocket.Conn
	Mu       sync.Mutex
	CLientID string
}

type Hub struct {
	Clients          map[*Client]bool
	Broadcast        chan []byte
	Register         chan *Client
	Unregister       chan *Client
	Mu               sync.RWMutex
	ProductStockRepo productstock.ProductStockRepositoryInterface
}

func NewHub(repo productstock.ProductStockRepositoryInterface) *Hub {
	return &Hub{
		Clients:          make(map[*Client]bool),
		Broadcast:        make(chan []byte),
		Register:         make(chan *Client),
		Unregister:       make(chan *Client),
		Mu:               sync.RWMutex{},
		ProductStockRepo: repo,
	}
}
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client] = true
			h.Mu.Unlock()
			log.Println("Registered new client with ID: ", client.CLientID)

			go h.checkAndNotifyLowStock()

		case client := <-h.Unregister:
			h.Mu.Lock()
			delete(h.Clients, client)
			h.Mu.Unlock()
			log.Println("Unregistered client with ID: ", client.CLientID)
		case message := <-h.Broadcast:
			h.Mu.RLock()
			for client := range h.Clients {
				go func(c *Client) {
					err := c.Conn.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						log.Println("Error writing to client: ", err)
						h.Unregister <- c
						c.Conn.Close()
					}
				}(client)
			}
			h.Mu.RUnlock()
		}
	}
}

func (h *Hub) checkAndNotifyLowStock() {
	products, err := h.ProductStockRepo.GetLowStockProducts()
	if err != nil {
		log.Println("Error getting low stock products: ", err)
		return
	}
	if len(products) > 0 {
		message := map[string]interface{}{
			"type":     "low_stock_notification",
			"message":  "Low stock products detacted",
			"products": products,
		}
		jsonMessage, err := json.Marshal(message)
		if err != nil {
			log.Println("Error marshalling message: ", err)
			return
		}
		h.Broadcast <- jsonMessage
	}

}

func WebSocketHandler(hub *Hub) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		client := &Client{Conn: c, CLientID: c.Query("clientId")}
		hub.Register <- client
		defer func() {
			hub.Unregister <- client
			c.Close()
		}()

		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}
			if messageType == websocket.TextMessage {
				log.Printf("Receivedmessage: %s", message)
				hub.Broadcast <- message
			}
		}
	})
}
