package router

import (
	"gin/handlers"
	"gin/middleware"
	"gin/repository"

	"github.com/gin-gonic/gin"
)

func SetupRouter(userRepo repository.UserRepository) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	userHandler := handlers.NewUserHandler(userRepo)
	authHandler := handlers.NewAuthHandler(userRepo)

	router.POST("login", authHandler.Login)
	router.GET("users/:id", userHandler.GetUser)
	router.GET("posts", middleware.AuthMiddleware(userRepo), userHandler.GetAuthUserPosts)

	return router
}
