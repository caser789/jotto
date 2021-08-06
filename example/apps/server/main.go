package main

import (
	"github.com/caser789/jotto/jotto"
	"github.com/caser789/jotto/sample"
)

func main() {
	app := motto.NewApplication(motto.HTTP, ":8080", sample.Routes)

	app.On(motto.BootEvent, sample.Boot)

	app.Boot()

	app.Run()
}
