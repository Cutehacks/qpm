// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"qpm.io/common"
	"qpm.io/qpm/core"
	"strconv"
	"strings"
	"text/template"
	"time"
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

	// TODO: Check ~/.gitconfig for name/email
	ic.Pkg.Author.Name = <-Prompt("Your name:", ic.Pkg.Author.Name)
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
	ic.Pkg.Version.Revision = "XXXXXXXX"

	ic.Pkg.Repository.Url = <-Prompt("Repository:", "")

	if err := ic.Pkg.Save(); err != nil {
		return err
	}

	bootstrap := <-Prompt("Generate boilerplate:", "Y/n")
	if len(bootstrap) == 0 || strings.ToLower(string(bootstrap[0])) == "y" {
		ic.GenerateBoilerplate()
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

func (ic InitCommand) WriteBoilerPlate(filename string, tpl *template.Template, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		ic.Error(err)
		return err
	}
	defer file.Close()

	err = tpl.Execute(file, data)
	if err != nil {
		ic.Error(err)
		return err
	}

	return nil
}

func (ic InitCommand) GenerateBoilerplate() {

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

	ic.WriteBoilerPlate(module.PriFile, modulePri, module)
	ic.WriteBoilerPlate(module.QrcFile, moduleQrc, module)
	ic.WriteBoilerPlate("qmldir", qmldir, module)
}

type License struct {
	Key       string
	Name      string
	Permitted []string
	Forbidden []string
	Body      string
}

func (ic *InitCommand) license(license string) error {

	url := "https://api.github.com/licenses/" + license
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github.drax-preview+json")

	var response *http.Response
	response, err = client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var info License
	body, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(body, &info)
	if err != nil {
		return err
	}

	// FIXME: this is probably tied to the MIT license layout
	info.Body = strings.Replace(info.Body, "[year]", strconv.Itoa(time.Now().Year()), -1)
	info.Body = strings.Replace(info.Body, "[fullname]", ic.Pkg.Author.Name, -1)

	var file *os.File
	file, err = os.Create(core.License)
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
