package main

import "gin/app"

func main() {
	app := app.NewApp()

	app.Router.Run("8090")
}
