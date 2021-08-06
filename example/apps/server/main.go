package main

import (
	"fmt"

	"github.com/caser789/jotto/example/common"
	"github.com/caser789/jotto/example/routes"
	"github.com/caser789/jotto/jotto"
)

func main() {
	cfg := common.LoadCfg("conf/conf.xml")
	app := motto.NewApplication(cfg, routes.Routes)

	app.SetContextFactory(common.ContextFactory)

	app.On(motto.BootEvent, common.Boot)
	app.Boot()

	fmt.Println(app.Run(nil))
}
