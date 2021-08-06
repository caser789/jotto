package main

import (
	"flag"
	"fmt"

	"github.com/caser789/jotto/jotto"
	"github.com/caser789/jotto/sample"
)

func main() {
	var protocol, address string

	flag.StringVar(&protocol, "protocol", motto.HTTP, "the protocol (HTTP or TCP)")
	flag.StringVar(&address, "address", ":8080", "the address to listen on")

	flag.Parse()

	app := motto.NewApplication(protocol, address, sample.Routes)

	app.On(motto.BootEvent, sample.Boot)
	app.Boot()

	fmt.Println(app.Run(nil))
}
