package jotto

import (
	"flag"
	"fmt"
	"sort"
	"strings"
)

// Command provides a unified interface for utilities that are supposed
// to be run via command-line.
type Command interface {

	// The command name by which the command should be invoked.
	// Namespacing is recommended for command names. For example,
	// you may create several commands related to 'transactions'
	// with names 'txn:audit', 'txn:generate-report', 'txn:analysis'.
	// Commands under the same namespace will be grouped together
	// when printed to the command line in the help message, making
	// commands easier to discover.
	Name() string

	// A short description of what the command does.
	Description() string

	// This is where you do your initialization, e.g.:
	//  - connect to external servers;
	//  - register command line flags.
	// Note, you don't need to run `flag.Parse()` here. Flags will
	// be parsed for you after `Boot()` is called.
	Boot(flagSet *flag.FlagSet) error

	// Run will be called when all initialization are done, i.e. after `Boot()`.
	//  - app: the bootstraped wallet server Application instance;
	//  - args: the arguments returned by (flag.Args())
	Run(app Application, args []string) error

	// This is where you do your cleanups, e.g.:
	//  - close connections;
	//  - close files.
	Shutdown() error
}

type BaseCommand struct{}

func (c *BaseCommand) Boot() error     { return nil }
func (c *BaseCommand) Shutdown() error { return nil }

// CommandBus is a repository where commands can be resigered and fetched.
type CommandBus struct {
	commands map[string]Command
}

// NewBus creates a new `CommandBus`
func NewCommandBus() *CommandBus {
	return &CommandBus{
		commands: make(map[string]Command),
	}
}

// Register registers a `Command` into the bus.
func (b *CommandBus) Register(command Command) {
	b.commands[command.Name()] = command
}

// Find finds a `Command` by `name` in the bus.
func (b *CommandBus) Find(name string) (cmd Command, err error) {
	cmd, ok := b.commands[name]

	if !ok {
		return nil, fmt.Errorf("Cannot find command: %s", name)
	}

	return
}

// Print prints a list of available commands currently registered in the bus.
func (b *CommandBus) Print() {
	fmt.Println("Available commands")

	groups := make(map[string][]Command)

	for name, cmd := range b.commands {
		parts := strings.Split(name, ":")

		namespace := parts[0]

		if len(parts) == 1 {
			namespace = ""
		}

		group, ok := groups[namespace]

		if !ok {
			group = []Command{}
			groups[namespace] = group
		}

		groups[namespace] = append(group, cmd)

	}

	namespaces := []string{}

	for namespace := range groups {
		namespaces = append(namespaces, namespace)
	}

	sort.Strings(namespaces)

	for _, namespace := range namespaces {
		indent := " "
		if namespace != "" {
			indent = "  "
			fmt.Printf("\n %s\n", namespace)
		}
		for _, cmd := range groups[namespace] {
			fmt.Printf("%s%s : %s\n", indent, cmd.Name(), cmd.Description())
		}
	}
}
