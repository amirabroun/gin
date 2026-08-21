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

type createGroupRoomRequest struct {
	Name      string `json:"name"`
	MemberIDs []uint `json:"member_ids"`
}

type markReadRequest struct {
	RoomID    uint `json:"room_id" binding:"required"`
	MessageID uint `json:"message_id" binding:"required"`
}

type markAllReadRequest struct {
	RoomID    uint `json:"room_id" binding:"required"`
	MessageID uint `json:"message_id" binding:"required"`
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

	if msg.RoomID == 0 {
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

	roomID, err := strconv.ParseUint(c.Param("roomID"), 10, 64)
	if err != nil || roomID == 0 {
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

func (h *MessageHandler) GetUserRooms(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	rooms, err := h.svc.GetUserRooms(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load rooms"})
		return
	}
	if rooms == nil {
		rooms = []entity.Room{}
	}

	c.JSON(http.StatusOK, rooms)
}

func (h *MessageHandler) CreateGroupRoom(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createGroupRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.svc.CreateGroupRoom(userID, req.MemberIDs, req.Name)
	if err != nil {
		switch err {
		case service.ErrInvalidRoomName:
			c.JSON(http.StatusBadRequest, gin.H{"error": "room name is required"})
		case service.ErrNoGroupMembers:
			c.JSON(http.StatusBadRequest, gin.H{"error": "member_ids is required"})
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "one or more users not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create group room"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"room_id": room.ID})
}

func (h *MessageHandler) MarkRead(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req markReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate room_id and message_id
	if req.RoomID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required and must be greater than 0"})
		return
	}
	if req.MessageID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required and must be greater than 0"})
		return
	}

	if err := h.svc.MarkRead(userID, req.RoomID, req.MessageID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotRoomMember):
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this room"})
			return
		case errors.Is(err, service.ErrInvalidMessage):
			c.JSON(http.StatusBadRequest, gin.H{"error": "message does not belong to this room"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark message as read"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "message marked as read",
	})
}

func (h *MessageHandler) GetReadStates(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	states, err := h.svc.GetReadStates(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load read states"})
		return
	}
	if states == nil {
		states = []entity.RoomState{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    states,
		"count":   len(states),
	})
}

func (h *MessageHandler) GetRoomReadStates(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Validate roomID parameter
	roomID, err := strconv.ParseUint(c.Param("roomID"), 10, 64)
	if err != nil || roomID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid room_id parameter",
			"hint":  "room_id must be a positive integer",
		})
		return
	}

	states, err := h.svc.GetRoomReadStates(userID, uint(roomID))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotRoomMember):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "forbidden",
				"hint":  "you are not a member of this room",
			})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to load room read states",
			})
			return
		}
	}
	if states == nil {
		states = []entity.RoomState{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    states,
		"count":   len(states),
		"room_id": roomID,
	})
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

func (h *MessageHandler) MarkAllRead(c *gin.Context) {
	userID, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req markAllReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate room_id
	if req.RoomID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required and must be greater than 0"})
		return
	}

	if err := h.svc.MarkAllRead(userID, req.RoomID, req.MessageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "all messages marked as read",
	})
}
