package app

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin/internal/core"
	"gin/internal/database"
	user "gin/internal/user"
)

type App struct {
	DB      *gorm.DB
	Router  *gin.Engine
	Modules []core.Module
}

func New() *App {
	return &App{
		Router: gin.Default(),
	}
}

func (app *App) RegisterDefaultDatabase() {
	app.DB = database.InitDB()
}

func (app *App) RegisterModules() {
	app.Modules = []core.Module{
		user.New(app.DB),
	}
}

func (app *App) RegisterRoutes() {
	app.Router.SetTrustedProxies(nil)

	for _, module := range app.Modules {
		module.RegisterRoutes(app.Router)
	}
}
