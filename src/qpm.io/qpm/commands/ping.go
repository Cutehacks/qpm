// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"golang.org/x/net/context"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
	"time"
)

type PingCommand struct {
	BaseCommand
}

func NewPingCommand(ctx core.Context) *PingCommand {
	return &PingCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (p PingCommand) Description() string {
	return "Pings the server"
}

func (p *PingCommand) RegisterFlags(flags *flag.FlagSet) {
}

func (p *PingCommand) Run() error {

	before := time.Now()

	_, err := p.Ctx.Client.Ping(context.Background(), &msg.PingRequest{})

	if err != nil {
		p.Fatal("Cannot ping server:" + err.Error())
	}

	d := time.Since(before)

	fmt.Printf("SUCCESS! Ping took %v\n", d)

	return nil
}
