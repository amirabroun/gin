package ws

import "encoding/json"

const (
	MessageTypeMessage  = "message"
	MessageTypePresence = "presence"
	MessageTypeTyping   = "typing"
)

type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	To      []uint          `json:"-"`
}

type message struct {
	ID       uint   `json:"id"`
	SenderID uint   `json:"sender_id"`
	RoomID   uint   `json:"room_id"`
	Content  string `json:"content"`
}

type presence struct {
	UserID uint `json:"user_id"`
	Online bool `json:"online"`
}

type typing struct {
	UserID  uint   `json:"user_id"`
	RoomID  uint   `json:"room_id"`
	Content string `json:"content"`
}

func PresenceEnvelope(userID uint, online bool) Envelope {
	return newEnvelope(MessageTypePresence, presence{
		UserID: userID,
		Online: online,
	})
}

func MessageEnvelope(senderID uint, roomID uint, content string, messageID uint) Envelope {
	return newEnvelope(MessageTypeMessage, message{
		ID:       messageID,
		SenderID: senderID,
		RoomID:   roomID,
		Content:  content,
	})
}

func TypingEnvelope(senderID uint, roomID uint, content string) Envelope {
	return newEnvelope(MessageTypeTyping, typing{
		UserID:  senderID,
		RoomID:  roomID,
		Content: content,
	})
}

func newEnvelope(messageType string, payload any) Envelope {
	raw, err := json.Marshal(payload)

	if err != nil {
		panic(err)
	}

	return Envelope{
		Type:    messageType,
		Payload: raw,
	}
}

func (e Envelope) ToUsers(userIDs []uint) Envelope {
	e.To = userIDs

	return e
}

func (e Envelope) ToUser(userID uint) Envelope {
	e.To = []uint{userID}

	return e
}

func (e Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
