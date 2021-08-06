package main

import (
	"fmt"

	"git.garena.com/caser789/jotto/jotto"
	"git.garena.com/caser789/jotto/sample"
	"git.garena.com/caser789/jotto/sample/commands"
)

func main() {
	bus := motto.NewCommandBus()
	bus.Register(commands.NewUpper())

	runner := motto.NewCliRunner(bus)

	app := motto.NewApplication(motto.HTTP, ":8080", sample.Routes)
	app.Boot()

	fmt.Println(app.Run(runner))
}
