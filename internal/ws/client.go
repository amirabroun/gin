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
	outboxCapacity = 256
)

type Client struct {
	hub         *Hub
	userID      uint
	contacts    []uint
	outbox      chan Envelope
	connection  *websocket.Conn
	closeOnce   sync.Once
	offlineOnce sync.Once
}

func NewClient(hub *Hub, userID uint, connection *websocket.Conn, contacts []uint) *Client {
	return &Client{
		hub:        hub,
		userID:     userID,
		contacts:   contacts,
		connection: connection,
		outbox:     make(chan Envelope, outboxCapacity),
	}
}

func (c *Client) Serve(onMessage func(*Client, []byte)) {
	c.hub.Register(c)

	go c.writeLoop()
	go c.readLoop(onMessage)
}

func (c *Client) UserID() uint { return c.userID }

func (c *Client) Contacts() []uint { return c.contacts }

func (c *Client) Send(envelope Envelope) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("client outbox is closed")
		}
	}()

	select {
		case c.outbox <- envelope:
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

func (c *Client) fail() {
	c.offlineOnce.Do(func() {
		c.hub.Unregister(c)
		c.Close()
	})
}

func (c *Client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.connection.Close()
	}()

	for {
		select {
			case envelope, ok := <-c.outbox:
				if !c.writeEnvelope(envelope, ok) {
					return
				}

			case <-ticker.C:
				if !c.writePing() {
					return
				}
		}
	}
}

func (c *Client) writeEnvelope(envelope Envelope, ok bool) bool {
	if !ok {
		c.writeMessage(websocket.CloseMessage, []byte{})
		return false
	}

	wire, err := envelope.Marshal()
	if err != nil {
		c.fail()
		return false
	}

	if err := c.writeMessage(websocket.TextMessage, wire); err != nil {
		c.fail()
		return false
	}

	return true
}

func (c *Client) writePing() bool {
	if err := c.writeMessage(websocket.PingMessage, nil); err != nil {
		c.fail()
		return false
	}

	return true
}

func (c *Client) writeMessage(messageType int, payload []byte) error {
	c.connection.SetWriteDeadline(time.Now().Add(writeWait))

	return c.connection.WriteMessage(messageType, payload)
}

func (c *Client) readLoop(onMessage func(*Client, []byte)) {
	defer c.connection.Close()
	c.setReadSettings()

	for {
		_, raw, err := c.connection.ReadMessage()
		if err != nil {
			c.handleReadError(err)
			c.fail()
			return
		}

		onMessage(c, raw)
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
