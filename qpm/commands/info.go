// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"os"
	"text/template"

	"golang.org/x/net/context"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
)

const infoBody = `
Name: {{.Package.Name}}
Author: {{.Package.Author.Name}} ({{.Package.Author.Email}})
Webpage: {{.Package.Webpage}}
License: {{.Package.License.String}}
Repository: {{.Package.Repository.Url}}
Description: {{.Package.Description}}
Dependencies:
{{- with .Dependencies}}
	{{range $index, $dependency := . }}
	{{$dependency.Name}}@{{$dependency.Version}}
	{{end}}
{{else}}
	None.
{{end}}
Versions:
{{- with .Versions}}
	{{range $index, $version := .}}
	{{- $date := $version.DatePublished | toDate -}}
	{{$version.Version.Label}} [{{$date.Format "02/01/06 15:04"}}]
	{{end -}}
{{else}}
	No versions have been published.
{{end}}
Installation Statistics:
{{- with .InstallStats}}
	Today: {{.Daily}}
	This week: {{.Weekly}}
	This month: {{.Monthly}}
	This year: {{.Yearly}}
	Total: {{.Total}}
{{else}}
	Not available.
{{end -}}
`

var funcs = template.FuncMap{
	"toDate": core.ToDateTime,
}

var infoTemplate = template.Must(template.New("info").Funcs(funcs).Parse(infoBody))

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

	if err := infoTemplate.Execute(os.Stdout, response); err != nil {
		p.Fatal(err.Error())
	}

	return nil
}
