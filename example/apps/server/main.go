package main

import (
	"flag"
	"fmt"

	"github.com/caser789/jotto/example/common"
	"github.com/caser789/jotto/example/routes"
	"github.com/caser789/jotto/jotto"
)

func main() {
	var protocol, address string

	flag.StringVar(&protocol, "protocol", motto.HTTP, "the protocol (HTTP or TCP)")
	flag.StringVar(&address, "address", ":8080", "the address to listen on")

	flag.Parse()

	app := motto.NewApplication(protocol, address, routes.Routes)

	app.On(motto.BootEvent, common.Boot)
	app.Boot()

	fmt.Println(app.Run(nil))
}
