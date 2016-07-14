// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"golang.org/x/net/context"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
)

type ListCommand struct {
	BaseCommand
}

func NewListCommand(ctx core.Context) *ListCommand {
	return &ListCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (sc ListCommand) Description() string {
	return "Lists all packages in the registry"
}

func (sc *ListCommand) RegisterFlags(flags *flag.FlagSet) {
}

func (sc *ListCommand) Run() error {

	req := &msg.ListRequest{}

	response, err := sc.Ctx.Client.List(context.Background(), req)
	if err != nil {
		sc.Fatal("ERROR:" + err.Error())
	}

	results := response.GetResults()
	core.PrintSearchResults(results)

	return nil
}
