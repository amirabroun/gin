package main

import "gin/app"

func main() {
	app := app.New()

	app.Router.Run("8090")
}
