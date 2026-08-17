package ws

import "sync"

type RoomMessage struct {
	MemberIDs []uint
	Payload   []byte
}

type Hub struct {
	connections map[uint]*Client
	register    chan *Client
	unregister  chan *Client
	roomMessage chan *RoomMessage
	mu          sync.RWMutex
}

func New() *Hub {
	return &Hub{
		connections: make(map[uint]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		roomMessage: make(chan *RoomMessage),
	}
}

func (h *Hub) Register(c *Client) { h.register <- c }

func (h *Hub) Unregister(c *Client) { h.unregister <- c }

func (h *Hub) PublishToRoom(roomMessage *RoomMessage) { h.roomMessage <- roomMessage }

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.removeClients(client)

		case roomMessage := <-h.roomMessage:
			h.handleRoomMessage(roomMessage)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[client.userID] = client
}

func (h *Hub) handleRoomMessage(roomMessage *RoomMessage) {
	slowClients := h.deliver(roomMessage)

	h.removeClients(slowClients...)
}

func (h *Hub) deliver(roomMessage *RoomMessage) []*Client {
	var slowClients []*Client

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userID := range roomMessage.MemberIDs {
		client, ok := h.connections[userID]
		if !ok {
			continue
		}

		err := client.Send(roomMessage.Payload)

		if err != nil {
			slowClients = append(slowClients, client)
		}
	}

	return slowClients
}

func (h *Hub) removeClients(clients ...*Client) {
	var toClose []*Client

	h.mu.Lock()
	for _, client := range clients {
		existing, ok := h.connections[client.userID]

		if ok && existing == client {
			delete(h.connections, client.userID)
			toClose = append(toClose, client)
		}
	}
	h.mu.Unlock()

	for _, client := range toClose {
		client.Close()
	}
}
