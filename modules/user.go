package user

import (
	"gin/handlers"
	"gin/repository"

	"gorm.io/gorm"
)

type Module struct {
	Repository *repository.UserRepository
	Handler    *handlers.UserHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewUserRepository(db)
	handler := handlers.NewUserHandler(repo)

	return &Module{
		Repository: repo,
		Handler: handler,
	}
}
