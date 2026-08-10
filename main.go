package main

import (
	"gin/database"
	"gin/repository"
	"gin/router"
)

func main() {
	db := database.InitDB()

	userRepo := repository.NewUserRepository(db)

	router.SetupRouter(userRepo).Run(":8090")
}
