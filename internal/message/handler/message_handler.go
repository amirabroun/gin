package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"gin/internal/message/entity"
	"gin/internal/message/service"
	"gin/internal/ws"
)

// upgrader handles the HTTP -> WebSocket upgrade.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type incomingMessage struct {
	RoomID  uint   `json:"room_id"`
	Content string `json:"content"`
}

type startDirectChatRequest struct {
	TargetUserID uint `json:"target_user_id"`
}

type MessageHandler struct {
	svc *service.MessageService
	hub *ws.Hub
}

func New(svc *service.MessageService, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{svc: svc, hub: hub}
}

func authUserID(c *gin.Context) (uint, bool) {
	v, exists := c.Get("auth_user_id")
	if !exists {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

func (h *MessageHandler) HandleWebSocket(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	contacts, err := h.svc.GetContactIDs(userID)
	if err != nil {
		contacts = nil
	}

	ws.NewClient(h.hub, userID, conn, contacts).Serve(h.handleIncomingMessage)
}

func (h *MessageHandler) handleIncomingMessage(client *ws.Client, raw []byte) {
	var env struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(raw, &env); err != nil {
		return
	}

	var msg incomingMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return
	}

	if msg.RoomID == 0 || msg.Content == "" {
		return
	}

	switch env.Type {
		case ws.MessageTypeMessage:
			h.svc.Send(client.UserID(), msg.RoomID, msg.Content)

		case ws.MessageTypeTyping:
			h.svc.PublishTyping(client.UserID(), msg.RoomID, msg.Content)
	}
}

func (h *MessageHandler) GetRoomMessages(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	messages, err := h.svc.GetRoomMessages(userID, uint(roomID), limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrNotRoomMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load room messages"})
		return
	}
	if messages == nil {
		messages = []entity.Message{}
	}

	c.JSON(http.StatusOK, messages)
}

func (h *MessageHandler) StartDirectChat(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req startDirectChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TargetUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_user_id is required"})
		return
	}

	room, err := h.svc.GetOrCreateDirectRoom(userID, req.TargetUserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
			return
		}
		if errors.Is(err, service.ErrSelfDirectChat) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot start a direct chat with yourself"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start direct chat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"room_id": room.ID})
}
