package main

import (
	"flag"
	"os"

	"git.garena.com/caser789/jotto/jotto"
	"git.garena.com/caser789/jotto/sample"
	"git.garena.com/caser789/jotto/sample/commands"
)

var bus *jotto.CommandBus

func help() {
	fmt.Printf("Usage: %s -<command-name> ...<flags> ...<args>    To run a command\n", os.Args[0])
	fmt.Printf("       %s -<command-name> -h                      To get usage information of a specific command\n\n", os.Args[0])
	bus.Print()
}

func main() {
	app := motto.NewApplication(motto.HTTP, ":8080", sample.Routes)
	app.On(motto.BootEvent, sample.Boot)

	bus = motto.NewCommandBus()
	bus.Register(commands.NewUpper())

	runner := motto.NewCliRunner(bus)

	runner.Attatch(app)
	runner.Run()
}
