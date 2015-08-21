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
	ic.Pkg.Name = <-Prompt("Unique package name:", cwd)
	ic.Pkg.Version.Label = <-Prompt("Initial version:", ic.Pkg.Version.Label)

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
	funcMap = template.FuncMap{
		"dotSlash": func(dots string) string {
			return strings.Replace(dots, ".", "/", -1)
		},
	}
	modulePri = template.Must(template.New("modulePri").Funcs(funcMap).Parse(`
RESOURCES += \
    $$PWD/{{.PackageName}}.qrc
`))
	moduleQrc = template.Must(template.New("moduleQrc").Funcs(funcMap).Parse(`
<RCC>
    <qresource prefix="{{dotSlash .Namespace}}/{{.Name}}">
        <file>qmldir</file>
    </qresource>
</RCC>
`))
	qmldir = template.Must(template.New("qmldir").Funcs(funcMap).Parse(`
module {{.Namespace}}.{{.Name}}
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
		PackageName string
		Namespace   string
		Name        string
	}{
		PackageName: ic.Pkg.Name,
		Namespace:   strings.ToLower(<-Prompt("Namespace:", "com.example")),
		Name:        strings.ToLower(<-Prompt("Module name:", ic.Pkg.Name)),
	}

	ic.WriteBoilerPlate(ic.Pkg.Name+".pri", modulePri, module)
	ic.WriteBoilerPlate(ic.Pkg.Name+".qrc", moduleQrc, module)
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
