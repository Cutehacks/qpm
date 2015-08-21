// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"qpm.io/common"
	"qpm.io/qpm/core"
	"strings"
)

type CheckCommand struct {
	BaseCommand
	pkg common.PackageWrapper
}

func NewCheckCommand(ctx core.Context) *CheckCommand {
	return &CheckCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (c CheckCommand) Description() string {
	return "Checks the package for common errors"
}

func (c *CheckCommand) RegisterFlags(flags *flag.FlagSet) {
}

func (c *CheckCommand) Run() error {

	// check the package.json file
	err := c.pkg.Load()
	if err != nil {
		c.Error(err)
		return err
	}

	c.pkg.Validate()

	// check the LICENSE file
	_, err = os.Stat(core.License)
	if err != nil {
		c.Error(err)
		return err
	}

	// check the .pri file
	_, err = os.Stat(c.pkg.Name + ".pri")
	if err != nil {
		c.Error(err)
		return err
	}

	// check the .qrc file
	_, err = os.Stat(c.pkg.Name + ".qrc")
	if err != nil {
		c.Error(err)
		return err
	}

	var prefix string
	prefix, err = c.qrc()
	if err != nil {
		c.Error(err)
		return err
	}
	if strings.Count(prefix, "/") <= 1 {
		c.Warning("the resource prefix is not a reverse domain name")
	}

	// check the qmldir file
	_, err = os.Stat("qmldir")
	if err != nil {
		c.Error(err)
		return err
	}
	var module string
	module, err = c.qmldir()

	if strings.Count(module, ".") <= 0 {
		c.Warning("the module name is not a reverse domain name")
	}

	if !strings.EqualFold(prefix, strings.Replace(module, ".", "/", -1)) {
		c.Log("ERROR: the resource prefix and module name do not match")
		return nil
	}

	fmt.Printf("OK!\n")

	return nil
}

type QRC_File struct {
	XMLName xml.Name `xml:"file"`
	Content string   `xml:",chardata"`
}

type QRC_Resource struct {
	XMLName xml.Name   `xml:"qresource"`
	Prefix  string     `xml:"prefix,attr"`
	Files   []QRC_File `xml:"file"`
}

type QRC_RCC struct {
	XMLName   xml.Name       `xml:"RCC"`
	Resources []QRC_Resource `xml:"qresource"`
}

func (c *CheckCommand) qrc() (prefix string, err error) {

	file, err := os.Open(c.pkg.Name + ".qrc")
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, _ := ioutil.ReadAll(file)

	var rcc QRC_RCC
	xml.Unmarshal(data, &rcc)

	if rcc.Resources != nil {
		return rcc.Resources[0].Prefix, nil
	}

	return "", err
}

func (c *CheckCommand) qmldir() (prefix string, err error) {

	file, err := os.Open("qmldir")
	if err != nil {
		return "", err
	}
	defer file.Close()

	const key = "module "
	const key_len = len(key)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		i := strings.Index(line, key)
		if i != -1 {
			return line[i+key_len : len(line)], nil
		}
	}

	return "", err
}
