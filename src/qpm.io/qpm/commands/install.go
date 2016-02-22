// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"os"
	"path/filepath"
	"qpm.io/common"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
	"qpm.io/qpm/vcs"
	"strings"
	"text/template"
)

var packageFuncs = template.FuncMap{
	"relPriFile": func(vendorDir string, dep *common.PackageWrapper) string {
		abs := filepath.Join(dep.RootDir(), dep.PriFile())
		rel, err := filepath.Rel(vendorDir, abs)
		if err == nil {
			return rel
		} else {
			return abs
		}
	},
	"relHeaderFile": func(vendorDir string, dep *common.PackageWrapper, headerFile string) string {
		abs := filepath.Join(dep.RootDir(), headerFile)
		rel, err := filepath.Rel(vendorDir, abs)
		if err == nil {
			return rel
		} else {
			return abs
		}
	},
}

var (
	// This template is very dense to avoid excessive whitespace in the generated code.
	// We can address this in a future version of Go (1.6?):
	// https://github.com/golang/go/commit/e6ee26a03b79d0e8b658463bdb29349ca68e1460
	vendorPri = template.Must(template.New("vendorPri").Funcs(packageFuncs).Parse(`
DEFINES += QPM_INIT\\(E\\)=\"E.addImportPath(QStringLiteral(\\\"qrc:/\\\"));\"
DEFINES += QPM_USE_NS
INCLUDEPATH += $$PWD
QML_IMPORT_PATH += $$PWD
HEADERS += $$PWD/vendor.h
{{$vendirDir := .VendorDir}}
{{range $dep := .Dependencies}}
include($$PWD/{{relPriFile $vendirDir $dep}}){{end}}
`))
)

var (
	// This template is very dense to avoid excessive whitespace in the generated code.
	// We can address this in a future version of Go (1.6?):
	// https://github.com/golang/go/commit/e6ee26a03b79d0e8b658463bdb29349ca68e1460
	vendorHeader = template.Must(template.New("vendorHeader").Funcs(packageFuncs).Parse(`
#include <QQmlEngine>
#include <QCoreApplication>

{{$vendirDir := .VendorDir}}
{{range $dep := .Dependencies}}
{{range $header := $dep.Package.Headers}}
#include "{{relHeaderFile $vendirDir $dep $header}}"
{{end}}
{{end}}

namespace qpm {

void init(const QCoreApplication &app, QQmlEngine &engine) {
    // Add qml components
    engine.addImportPath(QStringLiteral("qrc:/"));

    // Add C++ components
    {{range $dep := .Dependencies}}
    {{range $plugin := $dep.Package.Plugins}}
    // {{$plugin}}
    {{$plugin.Class}} pluginInstance_{{$plugin.Class}};
    pluginInstance_{{$plugin.Class}}.registerTypes("{{$plugin.Uri}}");
    pluginInstance_{{$plugin.Class}}.initializeEngine(&engine, "{{$plugin.Uri}}");
    {{end}}
    {{end}}
}

}
`))
)

type ProgressProxyReader struct {
	io.Reader
	total    int64
	length   int64
	progress float64
}

func (r *ProgressProxyReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if n > 0 {
		r.total += int64(n)
		percentage := float64(r.total) / float64(r.length) * float64(100)
		i := int(percentage / float64(10))
		is := fmt.Sprintf("%v", i)
		if percentage-r.progress > 2 {
			fmt.Fprintf(os.Stderr, is)
			r.progress = percentage
		}
	}
	return n, err
}

type InstallCommand struct {
	BaseCommand
	pkg       *common.PackageWrapper
	fs        *flag.FlagSet
	vendorDir string
}

func NewInstallCommand(ctx core.Context) *InstallCommand {
	return &InstallCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (i InstallCommand) Description() string {
	return "Installs a new package"
}

func (i *InstallCommand) RegisterFlags(flags *flag.FlagSet) {
	i.fs = flags

	// TODO: Support other directory names on the command line?
	var err error
	i.vendorDir, err = filepath.Abs(core.Vendor)
	if err != nil {
		i.vendorDir = core.Vendor
	}
}

func (i *InstallCommand) Run() error {

	packageName := i.fs.Arg(0)

	var err error
	i.pkg, err = common.LoadPackage("")
	if err != nil {
		// A missing package file is only an error if packageName is empty
		if os.IsNotExist(err) {
			if packageName == "" {
				err = fmt.Errorf("No %s file found", core.PackageFile)
				i.Error(err)
				return err
			} else {
				// Create a new package
				file, err := filepath.Abs(core.PackageFile)
				if err != nil {
					i.Error(err)
					return err
				}
				i.pkg = common.NewPackageWrapper(file)
			}
		} else {
			i.Error(err)
			return err
		}
	}

	var packageNames []string
	if packageName == "" {
		packageNames = i.pkg.Dependencies
	} else {
		packageNames = []string{packageName}
	}

	// Get list of dependencies from the server
	response, err := i.Ctx.Client.GetDependencies(context.Background(), &msg.DependencyRequest{
		packageNames,
		i.pkg.License,
	})
	if err != nil {
		i.Error(err)
		return err
	}

	if len(response.Dependencies) == 0 {
		i.Info("No package(s) found")
		return nil
	}

	// Show info, warnings, errors and address prompts before continuing
	for _, msg := range response.Messages {
		fmt.Printf("%s: %s\n", msg.Type.String(), msg.Title)

		if msg.Body != "" {
			fmt.Println(msg.Body)
		}

		if msg.Prompt {
			continueAnyway := <-Prompt("Continue anyway?", "Y/n")
			if len(continueAnyway) == 0 || strings.ToLower(string(continueAnyway[0])) == "y" {
				continue
			} else {
				return fmt.Errorf("Installation aborted.")
			}
		}
	}

	// create the vendor directory if needed
	if _, err = os.Stat(i.vendorDir); err != nil {
		err = os.Mkdir(i.vendorDir, 0755)
	}

	// Download and extract the packages
	packages := []*common.PackageWrapper{}
	for _, d := range response.Dependencies {
		p, err := i.install(d)
		if err != nil {
			return err
		}
		packages = append(packages, p)
	}

	// Save the dependencies in the package file
	err = i.save(packages)
	// FIXME: should we continue installing ?
	if err != nil {
		return err
	}

	err = i.postInstall()
	// FIXME: should we continue installing ?
	if err != nil {
		return err
	}

	return nil
}

func (i *InstallCommand) install(d *msg.Dependency) (*common.PackageWrapper, error) {

	signature := strings.Join([]string{d.Name, d.Version.Label}, "@")
	fmt.Println("Installing", signature)

	pkg, err := vcs.Install(d, i.vendorDir)
	if err != nil {
		i.Error(err)
		return nil, err
	}

	return pkg, nil
}

func (i *InstallCommand) save(newDeps []*common.PackageWrapper) error {

	existingDeps := i.pkg.ParseDependencies()

	for _, d := range newDeps {
		existingVersion, exists := existingDeps[d.Name]
		if exists {
			if d.Version.Label != existingVersion {
				existingSignature := strings.Join([]string{d.Name, existingVersion}, "@")
				message := fmt.Sprint(existingSignature, " is already a dependency. Replacing with version ", d.Version.Label, ".")
				i.Warning(message)
				for n, e := range i.pkg.Dependencies {
					if existingSignature == e {
						i.pkg.Dependencies[n] = d.GetDependencySignature()
						break
					}
				}
			}
		} else {
			i.pkg.Dependencies = append(i.pkg.Dependencies, d.GetDependencySignature())
		}
	}

	err := i.pkg.Save()
	if err != nil {
		i.Error(err)
		return err
	}

	return nil
}

func (i *InstallCommand) postInstall() error {
	if err := GenerateVendorPri(i.vendorDir, i.pkg); err != nil {
		i.Error(err)
		return err
	}
	if err := GenerateVendorHeader(i.vendorDir, i.pkg); err != nil {
		i.Error(err)
		return err
	}
	return nil
}

func GenerateVendorHeader(vendorDir string, pkg *common.PackageWrapper) error {
	depMap, err := common.LoadPackages(vendorDir)
	if err != nil {
		return err
	}

	var deps []*common.PackageWrapper
	for _, dep := range depMap {
		deps = append(deps, dep)
	}

	vendorHeaderFile := filepath.Join(vendorDir, core.Vendor+".h")

	data := struct {
		VendorDir    string
		Package      *common.PackageWrapper
		Dependencies []*common.PackageWrapper
	}{
		vendorDir,
		pkg,
		deps,
	}
	return core.WriteTemplate(vendorHeaderFile, vendorHeader, data)
}

// Generates a vendor.pri inside vendorDir using the information contained in the package file
// and the dependencies
func GenerateVendorPri(vendorDir string, pkg *common.PackageWrapper) error {
	depMap, err := common.LoadPackages(vendorDir)
	if err != nil {
		return err
	}

	var deps []*common.PackageWrapper
	for _, dep := range depMap {
		deps = append(deps, dep)
	}

	vendorPriFile := filepath.Join(vendorDir, core.Vendor+".pri")

	data := struct {
		VendorDir    string
		Package      *common.PackageWrapper
		Dependencies []*common.PackageWrapper
	}{
		vendorDir,
		pkg,
		deps,
	}

	return core.WriteTemplate(vendorPriFile, vendorPri, data)
}
