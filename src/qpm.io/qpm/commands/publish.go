// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"strings"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"qpm.io/common"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
	"qpm.io/qpm/vcs"
)

type PublishCommand struct {
	BaseCommand
	PackageName string
}

func NewPublishCommand(ctx core.Context) *PublishCommand {
	return &PublishCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (p PublishCommand) Description() string {
	return "Publishes a new module"
}

func (p *PublishCommand) RegisterFlags(flags *flag.FlagSet) {

}

func get(name string, echoOff bool) string {
	var val string
	for {
		if echoOff {
			val = <-PromptPassword(name + ":")
		} else {
			val = <-Prompt(name+":", "")
		}
		if val == "" {
			fmt.Printf("ERROR: Must enter a %s\n", name)
		} else {
			break
		}
	}
	return val
}

func LoginPrompt(ctx context.Context, client msg.QpmClient) (string, error) {

	email := get("email", false)
	password := get("password", true)

	loginRequest := &msg.LoginRequest{
		Email:    email,
		Password: password,
		Create:   false,
	}

	loginResp, err := client.Login(context.Background(), loginRequest)

	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			fmt.Println("User not found. Confirm password to create a new user.")
			confirm := get("password", true)
			if password != confirm {
				return "", fmt.Errorf("Passwords do not match.")
			}

			loginRequest.Create = true
			if loginResp, err = client.Login(context.Background(), loginRequest); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return loginResp.Token, nil
}

func (p *PublishCommand) Run() error {

	token, err := LoginPrompt(context.Background(), p.Ctx.Client)

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return err
	}

	fmt.Println("Running check")
	if err := NewCheckCommand(p.Ctx).Run(); err != nil {
		p.Fatal(err.Error())
	}

	wrapper, err := common.LoadPackage("")

	if err != nil {
		p.Fatal("Cannot read " + core.PackageFile + ": " + err.Error())
	}

	wrapper.Version.Revision, err = vcs.LastCommitSHA1()

	if err != nil {
		p.Fatal("Cannot get the last commit SHA1: " + err.Error())
	}

	if err := vcs.ValidateCommit(wrapper.Version.Revision); err != nil {
		p.Fatal(err.Error())
	}

	fmt.Println("Publishing")
	_, err = p.Ctx.Client.Publish(context.Background(), &msg.PublishRequest{
		Token:              token,
		PackageDescription: wrapper.Package,
	})

	if err != nil {
		p.Fatal("ERROR:" + err.Error())
	}

	tag := <-Prompt("Tag release:", "Y/n")
	if len(tag) == 0 || strings.ToLower(string(tag[0])) == "y" {
		vcs.Tag("qpm/" + wrapper.Version.Label)
	}

	fmt.Println("SUCCESS!")
	signature := strings.Join([]string{wrapper.Name, wrapper.Version.Label}, "@")
	fmt.Println("Publised package: " + signature)
	fmt.Println("Revision: " + wrapper.Version.Revision)

	return nil
}
