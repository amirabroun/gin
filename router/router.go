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

	router.GET("users/:id", userHandler.GetUser)
	router.GET("posts", middleware.AuthMiddleware(userRepo), userHandler.GetAuthUserPosts)

	router.Run(":8090")

	return router
}
