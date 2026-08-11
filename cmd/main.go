package main

import "gin/internal/app"

func main() {
	app := app.New()

	app.Router.Run(":8090")
}
