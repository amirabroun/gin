package ws

type RoomMessage struct {
	MemberIDs []uint
	Payload   []byte
}

// Hub state (connections map) is owned exclusively by the Run goroutine.
// All mutations must happen inside Run's select loop. Public methods
// (Register, Unregister, PublishToRoom) only send on channels — they
// must NEVER touch h.connections directly. If you're tempted to add
// a method that reads/writes h.connections, it belongs inside Run.
type Hub struct {
	connections map[uint]*Client
	register    chan *Client
	unregister  chan *Client
	roomMessage chan *RoomMessage
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
			h.disconnect(client)

		case roomMessage := <-h.roomMessage:
			h.sendAndPrune(roomMessage.MemberIDs, roomMessage.Payload)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.connections[client.userID] = client

	h.broadcastPresence(client.userID, client.Contacts(), true)
	h.sendOnlineContactsTo(client)
}

func (h *Hub) sendOnlineContactsTo(client *Client) {
	for _, contactID := range client.Contacts() {
		_, ok := h.connections[contactID]

		if !ok {
			continue
		}

		client.Send(presenceEnvelope(contactID, true))
	}
}

func (h *Hub) sendToUsers(userIDs []uint, payload []byte) []*Client {
	var slow []*Client

	for _, userID := range userIDs {
		client, ok := h.connections[userID]
		if !ok {
			continue
		}

		if err := client.Send(payload); err != nil {
			slow = append(slow, client)
		}
	}

	return slow
}

func (h *Hub) disconnect(clients ...*Client) {
	for _, client := range clients {
		if h.connections[client.userID] != client {
			continue
		}

		delete(h.connections, client.userID)
		h.broadcastPresence(client.userID, client.Contacts(), false)
		client.Close()
	}
}

func (h *Hub) broadcastPresence(userID uint, contactIDs []uint, online bool) {
	h.sendAndPrune(contactIDs, presenceEnvelope(userID, online))
}

func (h *Hub) sendAndPrune(userIDs []uint, payload []byte) {
	h.disconnect(h.sendToUsers(userIDs, payload)...)
}

func presenceEnvelope(userID uint, online bool) []byte {
	return MustEnvelope(MessageTypePresence, PresencePayload{
		UserID: userID,
		Online: online,
	})
}
