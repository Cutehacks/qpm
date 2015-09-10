// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"golang.org/x/net/context"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
)

type SearchCommand struct {
	BaseCommand
	fs *flag.FlagSet
}

func NewSearchCommand(ctx core.Context) *SearchCommand {
	return &SearchCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (sc SearchCommand) Description() string {
	return "Searches for packages containing the given string"
}

func (sc *SearchCommand) RegisterFlags(flags *flag.FlagSet) {
	sc.fs = flags
}

func (sc *SearchCommand) Run() error {

	packageName := sc.fs.Arg(0)

	req := &msg.SearchRequest{PackageName: packageName}

	response, err := sc.Ctx.Client.Search(context.Background(), req)
	if err != nil {
		sc.Fatal("ERROR:" + err.Error())
	}

	results := response.GetResults()
	core.PrintSearchResults(results)

	return nil
}
