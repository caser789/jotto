package main

import (
	"flag"
	"fmt"

	"git.garena.com/caser789/jotto/example/commands"
	"git.garena.com/caser789/jotto/example/common"
	"git.garena.com/caser789/jotto/example/routes"
	"git.garena.com/caser789/jotto/jotto"
)

func main() {
	bus := motto.NewCommandBus()
	bus.Register(commands.NewUpper())

	runner := motto.NewCliRunner(bus)

	var recipe string
	flag.StringVar(&recipe, "recipe", "conf/conf.xml", "The configuration file")

	cfg := common.LoadCfg("conf/conf.xml")
	app := motto.NewApplication(cfg, routes.Routes)
	app.Boot()

	fmt.Println(app.Run(runner))
}
