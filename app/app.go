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
	Modules Modules
}

type Modules struct {
	User *user.Module
}

func New() *App {
	app := &App{
		DB:     database.InitDB(),
		Router: gin.Default(),
	}

	app.Router.SetTrustedProxies(nil)

	app.registerModules()

	app.registerRoutes()

	return app
}

func (app *App) registerModules() *App {
	app.Modules.User = user.New(app.DB)

	return app
}

func (app *App) registerRoutes() *App {
	app.Router.POST("login", app.Modules.User.Handler.Login)
	app.Router.GET("users/:id", app.Modules.User.Handler.GetUser)
	app.Router.GET("posts", app.Modules.User.Middleware.RequireAuth(), app.Modules.User.Handler.GetAuthUserPosts)

	return app
}
