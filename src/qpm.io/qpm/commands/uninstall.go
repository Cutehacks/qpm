// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"qpm.io/common"
	"qpm.io/qpm/core"
)

type UninstallCommand struct {
	BaseCommand
	fs        *flag.FlagSet
	vendorDir string
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

	var err error
	u.vendorDir, err = filepath.Abs(core.Vendor)
	if err != nil {
		u.vendorDir = core.Vendor
	}
}

func (u *UninstallCommand) Run() error {

	packageName := u.fs.Arg(0)

	if packageName == "" {
		err := fmt.Errorf("Must supply a package to uninstall")
		u.Error(err)
		return err
	}

	dependencyMap, err := common.LoadPackages(u.vendorDir)
	if err != nil {
		u.Error(err)
		return err
	}

	toRemove, exists := dependencyMap[packageName]
	if !exists {
		err := fmt.Errorf("Package %s was not found", packageName)
		u.Error(err)
		return err
	}

	// Does the current directory contain a package file that needs updating?
	pkg, err := common.LoadPackage("")
	if err != nil && !os.IsNotExist(err) {
		u.Error(err)
		return err
	} else if err == nil {
		pkg.RemoveDependency(toRemove)
		if err := pkg.Save(); err != nil {
			u.Error(err)
			return err
		}
	}

	fmt.Println("Uninstalling", toRemove.Name)

	// Final step is to delete the dependency's directory. This should
	// be done last since after this step, the info about the package is
	// gone.
	if err := os.RemoveAll(toRemove.RootDir()); err != nil {
		u.Error(err)
		return err
	}

	// TODO: Cleanup empty leaf directories in parent dirs

	// Regenerate vendor.pri
	if err := GenerateVendorPri(u.vendorDir, pkg); err != nil {
		u.Error(err)
		return err
	}

	return nil
}
