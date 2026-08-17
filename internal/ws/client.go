package ws

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

type Client struct {
	userID     uint
	contacts   []uint
	outbox     chan []byte
	connection *websocket.Conn
	closeOnce  sync.Once
}

func NewClient(userID uint, connection *websocket.Conn) *Client {
	return &Client{
		userID:     userID,
		connection: connection,
		outbox:     make(chan []byte, maxMessageSize),
	}
}

func (c *Client) SetContacts(contactIDs []uint) {
	c.contacts = contactIDs
}

func (c *Client) Contacts() []uint {
	return c.contacts
}

func (c *Client) Send(message []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("client outbox is closed")
		}
	}()

	select {
	case c.outbox <- message:
		return nil

	default:
		return errors.New("client outbox is closed")
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.outbox)
	})
}

func (c *Client) WriteOnConnection() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.outbox:
			if !ok {
				c.writeMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.writeMessage(websocket.TextMessage, message)

			if err != nil {
				log.Printf("ws write error for user %d: %v", c.userID, err)
				return
			}

		case <-ticker.C:
			if err := c.writeMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("ws ping error for user %d: %v", c.userID, err)
				return
			}
		}
	}
}

func (c *Client) writeMessage(messageType int, payload []byte) error {
	c.connection.SetWriteDeadline(time.Now().Add(writeWait))

	return c.connection.WriteMessage(messageType, payload)
}

func (c *Client) ReadFromConnection(onMessage func([]byte), onDisconnect func(*Client)) {
	defer func() {
		onDisconnect(c)
		c.connection.Close()
	}()

	c.setReadSettings()

	for {
		_, message, err := c.connection.ReadMessage()
		if err != nil {
			c.handleReadError(err)
			return
		}

		onMessage(message)
	}
}

func (c *Client) setReadSettings() {
	c.connection.SetReadLimit(maxMessageSize)
	c.connection.SetReadDeadline(time.Now().Add(pongWait))
	c.connection.SetPongHandler(func(string) error {
		c.connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
}

func (c *Client) handleReadError(err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		log.Printf("ws unexpected close for user %d: %v", c.userID, err)
	}
}
