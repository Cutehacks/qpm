// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"golang.org/x/net/context"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
	"strings"
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

	if len(results) == 0 {
		fmt.Printf("No packages found.\n")
		return nil
	} else {
		fmt.Printf("\n%-40s %-20s\n", "Package", "Author")
		fmt.Printf("%s\n", strings.Repeat("-", 75))
	}

	for _, r := range results {
		fmt.Printf("%-40s %s\n",
			r.Name+"@"+r.Version,
			r.GetAuthor().Name+" <"+r.GetAuthor().Email+">",
		)
	}

	return nil
}
