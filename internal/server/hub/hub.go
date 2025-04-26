package hub

import "github.com/gorilla/websocket"

var (
	clients    = make(map[string]map[*Client]bool)
	broadcast  = make(chan Message)
	register   = make(chan *Client)
	unregister = make(chan *Client)
	stop       = make(chan struct{})
)

type Client struct {
	conn *websocket.Conn
	id   string
}

type Message struct {
	ID   string
	Data any
}

func init() {
	go RunHub()
}

func RunHub() {
	for {
		select {
		case client := <-register:
			if _, ok := clients[client.id]; !ok {
				clients[client.id] = make(map[*Client]bool)
			}
			clients[client.id][client] = true

		case client := <-unregister:
			if _, ok := clients[client.id]; ok {
				delete(clients[client.id], client)
			}

		case message := <-broadcast:
			if subscribers, ok := clients[message.ID]; ok {
				for client := range subscribers {
					client.conn.WriteJSON(message.Data)
				}
			}

		case <-stop:
			return
		}
	}
}

func StopHub() {
	close(stop)
}

func Register(id string, conn *websocket.Conn) *Client {
	client := &Client{conn: conn, id: id}
	register <- client
	return client
}

func (client *Client) Close() {
	unregister <- client
}

func Broadcast(id string, data any) {
	broadcast <- Message{ID: id, Data: data}
}
