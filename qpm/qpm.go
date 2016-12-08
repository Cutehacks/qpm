// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	cmd "qpm.io/qpm/commands"
	"qpm.io/qpm/core"
)

var registry *CommandRegistry

func Usage() {
	fmt.Println(`
qpm is a tool for managing Qt dependencies

Usage:
	qpm COMMAND [args]`)

	registry.CommandUsage()

	fmt.Printf("qpm@%s (built from %s)\n\n", core.Version, core.Build)
}

func main() {

	ctx := *core.NewContext()

	// Register new sub-commands here
	registry = NewCommandRegistry()
	registry.RegisterSubCommand("ping", cmd.NewPingCommand(ctx))
	registry.RegisterSubCommand("init", cmd.NewInitCommand(ctx))
	registry.RegisterSubCommand("search", cmd.NewSearchCommand(ctx))
	registry.RegisterSubCommand("list", cmd.NewListCommand(ctx))
	registry.RegisterSubCommand("info", cmd.NewInfoCommand(ctx))
	registry.RegisterSubCommand("install", cmd.NewInstallCommand(ctx))
	registry.RegisterSubCommand("uninstall", cmd.NewUninstallCommand(ctx))
	registry.RegisterSubCommand("publish", cmd.NewPublishCommand(ctx))
	registry.RegisterSubCommand("help", cmd.NewHelpCommand(ctx))
	registry.RegisterSubCommand("check", cmd.NewCheckCommand(ctx))
	registry.RegisterSubCommand("sign", cmd.NewSignCommand(ctx))
	registry.RegisterSubCommand("verify", cmd.NewVerifyCommand(ctx))
	//registry.RegisterSubCommand("deprecate", cmd.NewDeprecateCommand(ctx))
	//registry.RegisterSubCommand("prune", cmd.NewPruneCommand(ctx))

	if len(os.Args) < 2 {
		Usage()
		return
	}

	subCmd := os.Args[1]

	if !registry.Exists(subCmd) {
		Usage()
		return
	}

	command := registry.Get(subCmd)

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.Usage = func() {
		Usage()
		fs.PrintDefaults()
	}

	command.RegisterFlags(fs)

	fs.Parse(os.Args[2:])

	command.Run()
}
