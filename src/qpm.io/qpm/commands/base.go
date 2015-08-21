// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"qpm.io/qpm/core"
)

type BaseCommand struct {
	Ctx core.Context
}

func (bc BaseCommand) Log(msg string) {
	bc.Ctx.Log.Print(msg)
}

func (bc BaseCommand) Info(msg string) {
	bc.Ctx.Log.Print("INFO: " + msg)
}

func (bc BaseCommand) Warning(msg string) {
	bc.Ctx.Log.Print("WARNING: " + msg)
}

func (bc BaseCommand) Error(err error) {
	bc.Ctx.Log.Print("ERROR: " + err.Error())
}

func (bc BaseCommand) Fatal(msg string) {
	bc.Ctx.Log.Fatal(msg)
}
