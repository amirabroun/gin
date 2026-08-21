package message

import (
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"gin/internal/message/entity"
	"gin/internal/message/handler"
	"gin/internal/message/repository"
	"gin/internal/message/service"
	"gin/internal/ws"
)

type Module struct {
	Hub        *ws.Hub
	Repository repository.MessageRepository
	Service    *service.MessageService
	Handler    *handler.MessageHandler
}

func New(db *gorm.DB) *Module {
	if err := db.AutoMigrate(&entity.Room{}, &entity.RoomMember{}, &entity.Message{}, &entity.RoomState{}); err != nil {
		panic(err)
	}

	hub := ws.New()
	repo := repository.New(db)
	svc := service.New(repo, hub)
	h := handler.New(svc, hub)

	return &Module{
		Hub:        hub,
		Repository: repo,
		Service:    svc,
		Handler:    h,
	}
}

func (m *Module) RegisterRoutes(router *gin.Engine, requireAuth gin.HandlerFunc) {
	messages := router.Group("messages", requireAuth)
	{
		messages.POST("direct", m.Handler.StartDirectChat)
		messages.POST("groups", m.Handler.CreateGroupRoom)
		messages.GET("rooms", m.Handler.GetUserRooms)
		messages.GET("rooms/:roomID", m.Handler.GetRoomMessages)
	
		messages.GET("rooms/:roomID/read-states", m.Handler.GetRoomReadStates)
		messages.POST("read", m.Handler.MarkRead)
		messages.GET("read", m.Handler.GetReadStates)
		messages.POST("all-read", m.Handler.MarkAllRead)
	}

	router.GET("ws", requireAuth, m.Handler.HandleWebSocket)
}

func (m *Module) Run() {
	go m.Hub.Run()
}
