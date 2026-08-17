package ws

import "encoding/json"

const (
	MessageTypeMessage  = "message"
	MessageTypePresence = "presence"
)

type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type PresencePayload struct {
	UserID uint `json:"user_id"`
	Online bool `json:"online"`
}

func NewEnvelope(messageType string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return json.Marshal(Envelope{
		Type:    messageType,
		Payload: raw,
	})
}

func MustEnvelope(messageType string, payload any) []byte {
	data, err := NewEnvelope(messageType, payload)
	if err != nil {
		panic(err)
	}
	return data
}