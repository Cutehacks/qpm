// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"golang.org/x/net/context"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
)

type InfoCommand struct {
	BaseCommand
	fs *flag.FlagSet
}

func NewInfoCommand(ctx core.Context) *InfoCommand {
	return &InfoCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (p InfoCommand) Description() string {
	return "Displays information about the specified package"
}

func (p *InfoCommand) RegisterFlags(flags *flag.FlagSet) {
	p.fs = flags
}

func (p *InfoCommand) Run() error {

	packageName := p.fs.Arg(0)

	response, err := p.Ctx.Client.Info(context.Background(), &msg.InfoRequest{PackageName: packageName})

	if err != nil {
		p.Fatal(err.Error())
	}

	fmt.Printf("\nName: %s", response.Package.Name)
	fmt.Printf("\nAuthor: %s (%s)", response.Package.Author.Name, response.Package.Author.Email)
	fmt.Printf("\nWebpage: %s", response.Package.Webpage)
	fmt.Printf("\nLicense: %s", response.Package.License.String())
	fmt.Printf("\nRepository: %s", response.Package.Repository.Url)
	fmt.Printf("\nDescription: %s\n\n", response.Package.Description)

	return nil
}
