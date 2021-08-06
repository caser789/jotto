package main

import (
	"fmt"

	"github.com/caser789/jotto/jotto"
)

func boot(app interface{}) {
	a := app.(jotto.Application)

	fmt.Println("booted", a)
}

func main() {
	app := jotto.NewApplication(jotto.HTP, ":8080", Routes)

	app.On(jotto.BootEvent, boot)

	app.Boot()

	app.Run()
}
