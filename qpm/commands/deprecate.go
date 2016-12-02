// Copyright 2016 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"

	"golang.org/x/net/context"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
)

const dependentsBody = `
Dependencies:
{{- with .Dependencies}}
	{{range $index, $dependency := . }}
	{{$dependency.Name}}@{{$dependency.Version}}
	{{end}}
{{else}}
	None.
{{end}}
`

var dependentsTemplate = template.Must(template.New("dependents").Parse(dependentsBody))

type DeprecateCommand struct {
	BaseCommand
	fs *flag.FlagSet
}

func NewDeprecateCommand(ctx core.Context) *DeprecateCommand {
	return &DeprecateCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (d DeprecateCommand) Description() string {
	return "Deprecates the specified package"
}

func (d *DeprecateCommand) RegisterFlags(flags *flag.FlagSet) {
	d.fs = flags
}

func (d *DeprecateCommand) Dependents(packageName string) error {

	packageNames := []string{packageName}

	dependents, err := d.Ctx.Client.GetDependents(context.Background(), &msg.DependencyRequest{
		packageNames,
		msg.LicenseType_NONE,
	})
	if err != nil {
		return err
	}

	if err := dependentsTemplate.Execute(os.Stdout, dependents); err != nil {
		return err
	}

	return nil
}

func (d *DeprecateCommand) Run() error {

	packageName := d.fs.Arg(0)

	token, err := LoginPrompt(context.Background(), d.Ctx.Client)

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return err
	}

	if err := d.Dependents(packageName); err != nil {
		d.Fatal(err.Error())
	}

	sure := <-Prompt("Are you REALLY sure?:", "Y/n")
	if len(sure) == 0 || strings.ToLower(string(sure[0])) == "y" {
		_, err := d.Ctx.Client.Deprecate(context.Background(), &msg.DeprecateRequest{
			PackageName: packageName,
			Token:       token,
		})
		if err != nil {
			d.Fatal("ERROR:" + err.Error())
		}
	}

	return nil
}
