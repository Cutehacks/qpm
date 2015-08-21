// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
)

type SubCommand interface {
	Run() error
	Description() string
	RegisterFlags(*flag.FlagSet)
}

type CommandRegistry struct {
	Commands map[string]SubCommand
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		Commands: make(map[string]SubCommand),
	}
}

func (c *CommandRegistry) RegisterSubCommand(command string, handler SubCommand) error {
	if c.Exists(command) {
		return fmt.Errorf("Command already exists: %s", command)
	}

	c.Commands[command] = handler
	return nil
}

func (c CommandRegistry) Exists(command string) bool {
	_, exists := c.Commands[command]
	return exists
}

func (c CommandRegistry) Get(command string) SubCommand {
	return c.Commands[command]
}

func (c CommandRegistry) CommandUsage() {
	fmt.Println(`
The currently supported commands are:
	`)
	for k, v := range c.Commands {
		t := "\t\t"
		l := 2 - (len(k) / 4)
		for i := 0; i < l; i++ {
			t += "\t"
		}
		fmt.Printf("\t%s%s%s\n", k, t, v.Description())
	}
	fmt.Println("")
}
