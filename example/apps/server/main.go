package main

import (
	"flag"
	"fmt"

	"github.com/caser789/jotto/example/common"
	"github.com/caser789/jotto/example/routes"
	"github.com/caser789/jotto/jotto"
)

func main() {
	var recipe string
	flag.StringVar(&recipe, "recipe", "conf/conf.xml", "The configuration file")
	flag.Parse()

	cfg := common.LoadCfg(recipe)
	app := motto.NewApplication(cfg, routes.Routes)

	app.SetContextFactory(common.ContextFactory)

	app.On(motto.BootEvent, common.Boot)
	app.Boot()

	fmt.Println(app.Run(nil))
}
