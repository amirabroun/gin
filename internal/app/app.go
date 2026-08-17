package app

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin/internal/database"
	"gin/internal/message"
	user "gin/internal/user"
)

type App struct {
	DB            *gorm.DB
	Router        *gin.Engine
	UserModule    *user.Module
	MessageModule *message.Module
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
	app.UserModule = user.New(app.DB)
	app.MessageModule = message.New(app.DB)
}

func (app *App) RegisterRoutes() {
	app.Router.SetTrustedProxies(nil)

	app.UserModule.RegisterRoutes(app.Router)

	requireAuth := app.UserModule.Middleware.RequireAuth()
	app.MessageModule.RegisterRoutes(app.Router, requireAuth)

	app.MessageModule.Run()
}