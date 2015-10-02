// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/howeyc/gopass"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"qpm.io/common"
	"qpm.io/qpm/core"
	"qpm.io/qpm/vcs"
	"qpm.io/qpm/license"
)

func Prompt(prompt string, def string) chan string {
	replyChannel := make(chan string, 1)

	if def == "" {
		fmt.Printf(prompt + " ")
	} else {
		fmt.Printf(prompt+" [%s] ", def)
	}

	in := bufio.NewReader(os.Stdin)
	answer, _ := in.ReadString('\n')

	answer = strings.TrimSpace(answer)

	if len(answer) > 0 {
		replyChannel <- answer
	} else {
		replyChannel <- def
	}

	return replyChannel
}

func PromptPassword(prompt string) chan string {
	replyChannel := make(chan string, 1)
	fmt.Printf(prompt + " ")
	replyChannel <- string(gopass.GetPasswd())
	return replyChannel
}

func extractReverseDomain(email string) string {
	emailParts := strings.Split(email, "@")
	if len(emailParts) != 2 {
		return ""
	}
	domainParts := strings.Split(emailParts[1], ".")
	for i, j := 0, len(domainParts)-1; i < j; i, j = i+1, j-1 {
		domainParts[i], domainParts[j] = domainParts[j], domainParts[i]
	}
	return strings.Join(domainParts, ".")
}

type InitCommand struct {
	BaseCommand
	Pkg *common.PackageWrapper
}

func NewInitCommand(ctx core.Context) *InitCommand {
	return &InitCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (ic InitCommand) Description() string {
	return "Initializes a new module in the current directory"
}

func (ic *InitCommand) RegisterFlags(flag *flag.FlagSet) {

}

func (ic *InitCommand) Run() error {

	ic.Pkg = &common.PackageWrapper{Package: common.NewPackage()}

	ic.Pkg.Author.Name, _ = vcs.LastCommitAuthorName()
	ic.Pkg.Author.Name, _ = <-Prompt("Your name:", ic.Pkg.Author.Name)

	ic.Pkg.Author.Email, _ = vcs.LastCommitEmail()
	ic.Pkg.Author.Email = <-Prompt("Your email:", ic.Pkg.Author.Email)

	cwd, err := os.Getwd()
	if err != nil {
		ic.Error(err)
		cwd = ""
	} else {
		cwd = filepath.Base(cwd)
	}

	suggestedName := extractReverseDomain(ic.Pkg.Author.Email) + "." + cwd

	ic.Pkg.Name = <-Prompt("Unique package name:", suggestedName)
	ic.Pkg.Version.Label = <-Prompt("Initial version:", ic.Pkg.Version.Label)

	ic.Pkg.Repository.Url, err = vcs.RepositorySubURL()
	if err != nil {
		ic.Error(err)
	}
	ic.Pkg.Repository.Url = <-Prompt("Repository:", ic.Pkg.Repository.Url)

	filename, _ := ic.findPriFile()
	if len(filename) == 0 {
		filename = ic.Pkg.PriFile()
	}
	ic.Pkg.PriFilename = <-Prompt("Package .pri file:", filename)

	if err := ic.Pkg.Save(); err != nil {
		ic.Error(err)
		return err
	}

	bootstrap := <-Prompt("Generate boilerplate:", "Y/n")
	if len(bootstrap) == 0 || strings.ToLower(string(bootstrap[0])) == "y" {
		if err := ic.GenerateBoilerplate(); err != nil {
			return err
		}
		ic.license("mit") // FIXME: add support for more licenses
	}
	return nil
}

var (
	modulePri = template.Must(template.New("modulePri").Parse(`
RESOURCES += \
    $$PWD/{{.QrcFile}}
`))
	moduleQrc = template.Must(template.New("moduleQrc").Parse(`
<RCC>
    <qresource prefix="{{.QrcPrefix}}">
        <file>qmldir</file>
    </qresource>
</RCC>
`))
	qmldir = template.Must(template.New("qmldir").Parse(`
module {{.Package.Name}}
`))
)

func (ic InitCommand) GenerateBoilerplate() error {

	module := struct {
		Package   *common.PackageWrapper
		PriFile   string
		QrcFile   string
		QrcPrefix string
	}{
		Package:   ic.Pkg,
		PriFile:   ic.Pkg.PriFile(),
		QrcFile:   ic.Pkg.QrcFile(),
		QrcPrefix: ic.Pkg.QrcPrefix(),
	}

	if err := core.WriteTemplate(module.PriFile, modulePri, module); err != nil {
		return err
	}
	if err := core.WriteTemplate(module.QrcFile, moduleQrc, module); err != nil {
		return err
	}
	if err := core.WriteTemplate("qmldir", qmldir, module); err != nil {
		return err
	}
	return nil
}

func (ic *InitCommand) license(identifier string) error {

	println("LICENSE FETCH")
	info, err := license.GetLicense(identifier, ic.Pkg)

	var file *os.File
	file, err = os.Create(core.LicenseFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(info.Body)
	if err != nil {
		return err
	}

	return nil
}

func (ic *InitCommand) findPriFile() (string, error) {
	dirname := "." + string(filepath.Separator)

	d, err := os.Open(dirname)
	if err != nil {
		return "", err
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".pri" {
				return file.Name(), nil
			}
		}
	}

	return "", nil
}