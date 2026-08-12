package user

import (
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"gin/internal/user/handler"
	"gin/internal/user/middleware"
	"gin/internal/user/repository"
	"gin/internal/user/service"
)

type Module struct {
	Repository repository.UserRepository
	Service    *service.UserService
	Handler    *handler.UserHandler
	Middleware *middleware.Middleware
}

func New(db *gorm.DB) *Module {
	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)
	authMiddleware := middleware.New(repo)

	return &Module{
		Repository: repo,
		Service:    svc,
		Handler:    h,
		Middleware: authMiddleware,
	}
}

func (m *Module) RegisterRoutes(router *gin.Engine) {
	router.POST("login", m.Handler.Login)
	router.GET("users/:id", m.Handler.GetUser)

	posts := router.Group("posts", m.Middleware.RequireAuth())
	{
		posts.POST("", m.Handler.StoreAuthUserPost)
		posts.GET("", m.Handler.GetAuthUserPosts)
	}
}
