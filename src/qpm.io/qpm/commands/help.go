// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"qpm.io/qpm/core"
)

type HelpCommand struct {
	BaseCommand
	fs *flag.FlagSet
}

func NewHelpCommand(ctx core.Context) *HelpCommand {
	return &HelpCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (h HelpCommand) Description() string {
	return "Shows the help text for a command"
}

func (h *HelpCommand) RegisterFlags(flags *flag.FlagSet) {
	h.fs = flags
}

func (h *HelpCommand) Run() error {

	commandName := h.fs.Arg(0)

	switch commandName {
	case "ping":
		fmt.Println(`
Checks if we are able to reach the server.

Usage:
	qpm ping
`)

	case "init":
		fmt.Println(`
Generates the necessary files for publishing a package to the qpm registry.

Usage:
	qpm init
`)

	case "install":
		fmt.Println(`
Installs the packages listed as dependencies in the package file or the given [PACKAGE].

Usage:
	qpm install [PACKAGE]
`)

	case "uninstall":
		fmt.Println(`
Removes the given [PACKAGE] from the project and deletes the associated files.

Usage:
	qpm uninstall [PACKAGE]
`)

	case "publish":
		fmt.Println(`
Publishes project as a package in the qpm registry.

Usage:
	qpm publish
`)

	case "sign":
		fmt.Println(`
Creates a PGP signature for contents of the project.

Usage:
	qpm sign
`)

	case "verify":
		fmt.Println(`
Verifies the the content and publisher of the given [PACKAGE], provided the package has been signed.

Usage:
	qpm verify [PACKAGE]
`)

	case "help":
		fallthrough

	default:
		fmt.Println(`
Shows the help text for the given [COMMAND]. If [COMMAND] is empty, it shows this text.

Usage:
	qpm help COMMAND
`)

	}

	return nil
}
