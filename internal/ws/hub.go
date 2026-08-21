package ws

// Hub state (connections map) is owned exclusively by the Run goroutine.
// All mutations must happen inside Run's select loop.
// Public methods (Register, Unregister, PublishToRoom) only send on channels
// They must NEVER touch h.connections directly.
// If you're tempted to add a method that reads/writes h.connections, it belongs inside Run.
type Hub struct {
	connections map[uint]*Client
	register    chan *Client
	unregister  chan *Client
	deliver     chan Envelope
}

func New() *Hub {
	return &Hub{
		connections: make(map[uint]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		deliver:     make(chan Envelope),
	}
}

func (h *Hub) Register(c *Client) { h.register <- c }

func (h *Hub) Unregister(c *Client) { h.unregister <- c }

func (h *Hub) Deliver(envelope Envelope) { h.deliver <- envelope }

func (h *Hub) Run() {
	for {
		select {
			case client := <-h.register:
				h.connect(client)

			case client := <-h.unregister:
				h.disconnect(client)

			case envelope := <-h.deliver:
				h.sendAndPrune(envelope)
		}
	}
}

func (h *Hub) connect(client *Client) {
	h.connections[client.userID] = client

	h.sendOnlineContactsTo(client)

	envelope := PresenceEnvelope(client.userID, true).ToUsers(client.Contacts())
	h.sendAndPrune(envelope)
}

func (h *Hub) disconnect(clients ...*Client) {
	for _, client := range clients {
		if h.connections[client.userID] != client {
			continue
		}

		delete(h.connections, client.userID)

		envelope := PresenceEnvelope(client.userID, false).ToUsers(client.Contacts())
		h.sendAndPrune(envelope)

		client.Close()
	}
}

func (h *Hub) sendAndPrune(envelope Envelope) {
	h.disconnect(h.send(envelope)...)
}

func (h *Hub) send(envelope Envelope) []*Client {
	var slow []*Client

	for _, userID := range envelope.To {
		client, ok := h.connections[userID]
		if !ok {
			continue
		}

		err := client.Send(envelope)

		if err != nil {
			slow = append(slow, client)
		}
	}

	return slow
}

func (h *Hub) sendOnlineContactsTo(client *Client) {
	for _, contactID := range client.Contacts() {
		_, ok := h.connections[contactID]

		if !ok {
			continue
		}

		envelope := PresenceEnvelope(contactID, true).ToUser(client.userID)

		h.send(envelope)
	}
}