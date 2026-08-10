package app

import (
	"gin/database"
	"gin/middleware"
	"gin/modules"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	DB      *gorm.DB
	Router  *gin.Engine
	Modules Modules
}

type Modules struct {
	User *user.Module
}

func NewApp() *App {
	db := database.InitDB()
	router := gin.Default()

	app := &App{
		DB:     db,
		Router: router,
	}

	app.Router.SetTrustedProxies(nil)

	app.Modules.User = user.NewModule(db)

	userRepo := app.Modules.User.Repository

	router.POST("login", app.Modules.User.Handler.Login)
	router.GET("users/:id", app.Modules.User.Handler.GetUser)
	router.GET("posts", middleware.AuthMiddleware(userRepo), app.Modules.User.Handler.GetAuthUserPosts)

	return app
}
