package main

import "gin/internal/app"

func main() {
	app := app.New()

	app.RegisterDefaultDatabase()
	app.RegisterModules()
	app.RegisterRoutes()

	app.Router.Run(":8090")
}
