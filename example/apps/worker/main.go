package main

import (
	"flag"

	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/duanzy/motto/sample/common"
	"git.garena.com/duanzy/motto/sample/jobs"
	"git.garena.com/duanzy/motto/sample/routes"
)

func main() {

	var recipe string
	var workers int

	flag.StringVar(&recipe, "recipe", "conf/conf.xml", "The configuration file")
	flag.IntVar(&workers, "workers", 100, "The workers pool size")
	flag.Parse()

	runner := motto.NewQueueWorkerRunner("default:main", workers)

	cfg := common.NewConfiguration(recipe)

	// Create application instance
	app := motto.NewApplication(cfg, routes.Routes, jobs.Jobs, runner)

	// Set logger and context factory
	app.SetLoggerFactory(common.NewCommonLogger)
	app.SetContextFactory(common.ContextFactory)

	// Register boot event listener
	app.On(motto.BootEvent, common.Boot)
	app.On(motto.ReloadEvent, common.Reload)
	app.On(motto.TerminateEvent, common.Terminate)

	soul := motto.NewSoul([]motto.Application{app})

	soul.Serve()
}
