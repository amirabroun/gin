package main

import (
	"github.com/gin-gonic/gin"

	"gin/database"
	"gin/handlers"
	"gin/repository"
)

func main() {
	db := database.InitDB()

	router := gin.Default()
	router.SetTrustedProxies(nil)

	userRepo := repository.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	router.GET("users/:id", userHandler.GetUser)

	router.Run(":8090")
}
