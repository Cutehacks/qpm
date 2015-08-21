// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"os"
	"qpm.io/common"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
	"strings"
)

type UninstallCommand struct {
	BaseCommand
	pkg common.PackageWrapper
	fs  *flag.FlagSet
}

func NewUninstallCommand(ctx core.Context) *UninstallCommand {
	return &UninstallCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (u UninstallCommand) Description() string {
	return "Uninstalls a package"
}

func (u *UninstallCommand) RegisterFlags(flags *flag.FlagSet) {
	u.fs = flags
}

func (u *UninstallCommand) Run() error {

	packageName := u.fs.Arg(0)

	if packageName == "" {
		return nil
	}

	err := u.pkg.Load()
	if err != nil {
		u.Error(err)
		return err
	}

	// remove the dependency
	i := u.index(len(u.pkg.Dependencies), func(i int) bool {
		signature := u.pkg.Dependencies[i]
		parts := strings.Split(signature, "@")
		return strings.ToLower(parts[0]) == strings.ToLower(packageName)
	})
	if i == -1 {
		fmt.Println("The package", packageName, "was not found")
		return nil
	}

	fmt.Println("Uninstalling", u.pkg.Dependencies[i])

	u.pkg.Dependencies = append(u.pkg.Dependencies[:i], u.pkg.Dependencies[i+1:]...)

	// Save the dependencies in package.json
	err = u.pkg.Save()
	if err != nil {
		u.Error(err)
		return err
	}

	var dependencies []*msg.Dependency
	for _, signature := range u.pkg.Dependencies {
		parts := strings.Split(signature, "@")
		var dep msg.Dependency
		dep.Name = strings.ToLower(parts[0])
		var ver msg.Package_Version
		ver.Label = strings.ToLower(parts[1])
		dep.Version = &ver
		dependencies = append(dependencies, &dep)
	}

	u.pkg.UpdatePri(dependencies)

	// Remove the package files
	os.RemoveAll(core.Vendor + "/" + packageName)

	return nil
}

func (u *UninstallCommand) index(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
