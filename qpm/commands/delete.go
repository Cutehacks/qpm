// Copyright 2016 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"

	"golang.org/x/net/context"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
)

type DeleteCommand struct {
	BaseCommand
	fs *flag.FlagSet
}

func NewDeleteCommand(ctx core.Context) *DeleteCommand {
	return &DeleteCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (d DeleteCommand) Description() string {
	return "Deletes the specified module"
}

func (d *DeleteCommand) RegisterFlags(flags *flag.FlagSet) {
	d.fs = flags
}

func (d *DeleteCommand) prompt(packageName string) error {
	fmt.Printf("Are you ABSOLUTELY sure?\nPlease type in the name of the package to confirm.\n")
	var val string
	for {
		val = <-Prompt("Package name:", "")
		if val != packageName {
			fmt.Printf("ERROR: Enter the package name or use <ctrl+c> to abort\n")
		} else {
			break
		}
	}
	return nil
}

func (d *DeleteCommand) Run() error {

	packageName := d.fs.Arg(0)

	token, err := LoginPrompt(context.Background(), d.Ctx.Client)

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return err
	}

	if err := NewDeprecateCommand(d.Ctx).Dependents(packageName); err != nil {
		d.Fatal(err.Error())
	}

	if err := d.prompt(packageName); err != nil {
		d.Fatal(err.Error())
	}

	_, err = d.Ctx.Client.Delete(context.Background(), &msg.DeleteRequest{
		packageName,
		token,
	})

	if err != nil {
		d.Fatal("ERROR:" + err.Error())
	}

	return nil
}
