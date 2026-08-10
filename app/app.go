package app

import (
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"gin/database"
	user "gin/modules"
)

type App struct {
	DB      *gorm.DB
	Router  *gin.Engine
	Modules *Modules
}

func New() *App {
	db := database.InitDB()

	app := &App{
		DB:      db,
		Router:  gin.Default(),
		Modules: NewModules(db),
	}

	app.Router.SetTrustedProxies(nil)
	app.registerRoutes()

	return app
}

func (app *App) registerRoutes() *App {
	app.Router.POST("login", app.Modules.User.Handler.Login)
	app.Router.GET("users/:id", app.Modules.User.Handler.GetUser)
	app.Router.GET("posts", app.Modules.User.Middleware.RequireAuth(), app.Modules.User.Handler.GetAuthUserPosts)

	return app
}

type Modules struct {
	User *user.Module
}

func NewModules(db *gorm.DB) *Modules {
	return &Modules{
		User: user.New(db),
	}
}
