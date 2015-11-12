package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
	"text/template"
	"time"
	"golang.org/x/net/context"
)

const (
	packageName     = "io.qpm.cli"
	stagingDir      = "packages"
	repositoryDir   = "repository"
	qpmLicense      = msg.LicenseType_ARTISTIC_2_0
	licenseAddendum = `
--------------------------

The following additional terms shall apply to use of the qpm software, the qpm.io
website and repository:

"qpm" and "qpm.io" are owned by Cutehacks AS. All rights reserved.

Modules published on the qpm registry are not officially endorsed by Cutehacks AS.

Data published to the qpm registry is not part of qpm itself, and is the sole
property of the publisher. While every effort is made to ensure accountability,
there is absolutely no guarantee, warranty, or assertion expressed or implied
as to the quality, fitness for a specific purpose, or lack of malice in any
given qpm package.  Packages downloaded through qpm.io are independently licensed
and are not covered by this license.
	`
)

var (
	platforms = map[string]string{
		"windows_386":   "qpm.exe",
		"windows_amd64": "qpm.exe",
		"linux_386":     "qpm",
		"linux_amd64":   "qpm",
		"darwin_386":    "qpm",
		"darwin_amd64":  "qpm",
	}

	rootPackageXml = template.Must(template.New("rootPackageXml").Parse(
		`<?xml version="1.0"?>
<Package>
    <DisplayName>qpm</DisplayName>
    <Description>qpm is a command line tool for installing packages from the qpm.io repository.</Description>
    <Version>{{.Version}}</Version>
    <ReleaseDate>{{.ReleaseDate}}</ReleaseDate>
    <Name>{{.Root}}</Name>
    <Licenses>
        <License name="License Agreement" file="license.txt" />
    </Licenses>
    <UpdateText>Initial release</UpdateText>
    <Default>true</Default>
    <ForcedInstallation>false</ForcedInstallation>
    <Essential>false</Essential>
</Package>
	`))

	platformPackageXml = template.Must(template.New("platformPackageXml").Parse(
		`<?xml version="1.0" encoding="UTF-8"?>
<Package>
    <DisplayName>{{.Platform}} binaries</DisplayName>
    <Description>qpm binaries for running on {{.Platform}}</Description>
    <Version>{{.Version}}</Version>
    <ReleaseDate>{{.ReleaseDate}}</ReleaseDate>
    <Name>{{.Root}}.{{.Platform}}</Name>
    <Default>false</Default>
</Package>
`))
)

type packageInfo struct {
	Root        string
	Platform    string
	ReleaseDate string
	Version     string
}

func newDir(dir string) string {
	if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil {
		log.Fatalf("could not create %s, %s", dir, err.Error())
	}
	return dir
}

func copyBinary(src, dest string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer d.Close()

	if _, err = io.Copy(d, s); err != nil {
		return err
	}
	if err := d.Sync(); err != nil {
		return err
	}

	return d.Chmod(0755)
}

func writeText(filename, text string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create file %s: %s", filename, err.Error())
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}

func writeTemplate(filename string, tpl *template.Template, pkg *packageInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create file %s: %s", filename, err.Error())
		return err
	}
	defer file.Close()

	err = tpl.Execute(file, pkg)
	if err != nil {
		log.Fatalf("Could not generate file %s: %s", filename, err.Error())
		return err
	}

	return nil
}

func main() {

	binDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("Error getting bin dir: %v\n", err)
	}

	var outputDir string
	if len(os.Args) < 2 {
		outputDir, err = filepath.Abs(repositoryDir)
	} else {
		outputDir, err = filepath.Abs(os.Args[1])
	}
	if err != nil {
		log.Fatalf("Error getting output dir: %v\n", err)
	}

	staging := newDir(filepath.Join(outputDir, stagingDir))
	log.Printf("Generated repository at %s\n", outputDir)

	buffer := bytes.NewBufferString(staging)
	reset := buffer.Len()

	pkg := &packageInfo{
		Root:        packageName,
		Version:     core.Version,
		ReleaseDate: time.Now().Format("2006-01-02"),
	}

	rootMetaDir := newDir(filepath.Join(staging, packageName, "meta"))
	newDir(filepath.Join(staging, packageName, "data"))

	if err = writeTemplate(filepath.Join(rootMetaDir, "package.xml"), rootPackageXml, pkg); err != nil {
		log.Fatalf("Could not generate package.xml for root %s", err.Error())
	}

	req := &msg.LicenseRequest{
		Package: &msg.Package{
			Name:        "qpm",
			Description: "A package manager for Qt",
			License: qpmLicense,
			Version: &msg.Package_Version{
				Label: core.Version,
			},
			Author: &msg.Package_Author{
				Name:  "Cutehacks AS",
			},
		},
	}

	ctx := core.NewContext()

	license, err := ctx.Client.GetLicense(context.Background(), req)
	if err != nil {
		log.Fatalf("Could not fetch license info:", err.Error())
	}

	licenseTxt := license.Body + licenseAddendum

	if err = writeText(filepath.Join(rootMetaDir, "license.txt"), licenseTxt); err != nil {
		log.Fatalf("Could not generate license.txt for root %s", err.Error())
	}

	for platform, binary := range platforms {
		srcDir := filepath.Join(binDir, platform)
		if _, err := os.Stat(srcDir); err != nil {
			log.Printf("Platform %s does not exist", platform)
			continue
		}

		// Create meta dir
		buffer.Truncate(reset)
		buffer.WriteString("/" + packageName + ".")
		buffer.WriteString(platform)
		buffer.WriteString("/meta/")
		metaDir := buffer.String()
		newDir(metaDir)

		// Create data dir
		buffer.Truncate(reset)
		buffer.WriteString("/" + packageName + ".")
		buffer.WriteString(platform)
		buffer.WriteString("/data/qpm/")
		dataDir := buffer.String()
		newDir(dataDir)

		if err := copyBinary(filepath.Join(srcDir, binary), filepath.Join(dataDir, binary)); err != nil {
			log.Fatalf("Cannot copy binary to %s: %s", dataDir, err.Error())
		}

		pkg.Platform = platform
		if err := writeTemplate(filepath.Join(metaDir, "package.xml"), platformPackageXml, pkg); err != nil {
			log.Fatalf("Could not generate package.xml for %s: %s", platform, err.Error())
		}
	}
}
