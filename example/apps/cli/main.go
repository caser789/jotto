package main

import (
	"flag"
	"fmt"
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
	bus = jotto.NewCommandBus()

	bus.Register(commands.NewUpper())

	if len(os.Args) < 2 {
		help()
		return
	}

	name := os.Args[1]

	// 2. Find command in the bus
	command, err := bus.Find(name[1:])

	if err != nil {
		help()
		return
	}

	flag.Bool(command.Name(), true, "command name")

	// 3. Run command initializations.
	command.Boot()

	flag.Parse()

	app := jotto.NewApplication(jotto.HTTP, ":8080", sample.Routes)
	app.On(jotto.BootEvent, sample.Boot)
	app.Boot()

	// 4. Run the command.
	command.Run(app, flag.Args())
}
