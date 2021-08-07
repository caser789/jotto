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

	cfg := common.NewConfiguration(recipe)

	// Create application instance
	app := motto.NewApplication(cfg, routes.Routes, nil)

	// Set logger and context factory
	app.SetLoggerFactory(common.NewCommonLogger)
	app.SetContextFactory(common.ContextFactory)

	// Register boot event listener
	app.On(motto.BootEvent, common.Boot)
	app.On(motto.ReloadEvent, common.Reload)

	// Boot the application
	app.Boot()

	// Start serving
	fmt.Println(app.Run(nil))
}
