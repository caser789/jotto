package main

import (
	"flag"
	"fmt"

	"git.garena.com/common/gocommon"
	"github.com/caser789/jotto/example/common"
	"github.com/caser789/jotto/example/routes"
	"github.com/caser789/jotto/jotto"
)

func main() {

	var recipe string
	flag.StringVar(&recipe, "recipe", "conf/conf.xml", "The configuration file")
	flag.Parse()

	cfg := common.LoadCfg(recipe)

	// Initialise the logger
	gocommon.LoggerInit("log/upper.log", 86400, 1000*1000*1000, 30, 3)

	// Create application instance
	app := motto.NewApplication(cfg, routes.Routes)

	// Set logger and context factory
	app.SetLoggerFactory(common.NewCommonLogger)
	app.SetContextFactory(common.ContextFactory)

	// Register boot event listener
	app.On(motto.BootEvent, common.Boot)

	// Boot the application
	app.Boot()

	// Start serving
	fmt.Println(app.Run(nil))
}
