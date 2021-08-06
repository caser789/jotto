package main

import (
	"fmt"

	"git.garena.com/caser789/jotto/example/commands"
	"git.garena.com/caser789/jotto/example/routes"
	"git.garena.com/caser789/jotto/jotto"
)

func main() {
	bus := motto.NewCommandBus()
	bus.Register(commands.NewUpper())

	runner := motto.NewCliRunner(bus)

	app := motto.NewApplication(motto.HTTP, ":8080", routes.Routes)
	app.Boot()

	fmt.Println(app.Run(runner))
}
