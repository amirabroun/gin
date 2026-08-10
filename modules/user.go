package user

import (
	"gin/handlers"
	"gin/middleware"
	"gin/repository"

	"gorm.io/gorm"
)

type Module struct {
	Repository repository.UserRepository
	Handler    *handlers.UserHandler
	Middleware *middleware.Middleware
}

func New(db *gorm.DB) *Module {
	repo := repository.New(db)
	handler := handlers.New(repo)
	authMiddleware := middleware.New(repo)

	return &Module{
		Repository: repo,
		Handler:    handler,
		Middleware: authMiddleware,
	}
}
