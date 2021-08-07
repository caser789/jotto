package main

import (
	"fmt"

	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/duanzy/motto/sample/commands"
	"git.garena.com/duanzy/motto/sample/common"
	"git.garena.com/duanzy/motto/sample/routes"
)

func main() {
	bus := motto.NewCommandBus()
	bus.Register(commands.NewUpper())
	bus.Register(commands.NewJob())
	bus.Register(commands.NewTest())
	bus.Register(commands.NewWait())

	runner := motto.NewCliRunner(bus)

	cfg := common.NewConfiguration("conf/conf.xml")
	app := motto.NewApplication(cfg, routes.Routes, nil, runner)
	app.Boot()

	fmt.Println(app.Run())
}
