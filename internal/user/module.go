package user

import (
	"gorm.io/gorm"

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
