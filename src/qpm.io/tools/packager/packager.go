package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

const (
	packageName   = "io.qpm.cli"
	stagingDir    = "packages"
	repositoryDir = "repository"
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

	licenseTxt = template.Must(template.New("licenseTxt").Parse(`
qpm is available under the terms of a license.
	`))

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
		log.Fatalf("Error getting output dirr: %v\n", err)
	}

	staging := newDir(outputDir + "/" + stagingDir + "/")
	log.Printf("Generated repository at %s\n", outputDir)

	buffer := bytes.NewBufferString(staging)
	reset := buffer.Len()

	pkg := &packageInfo{
		Root:        packageName,
		Version:     "0.0.1",
		ReleaseDate: time.Now().Format("2006-01-02"),
	}

	rootMetaDir := newDir(staging + packageName + "/meta/")
	newDir(staging + packageName + "/data")

	if err = writeTemplate(rootMetaDir+"package.xml", rootPackageXml, pkg); err != nil {
		log.Fatalf("Could not generate package.xml for root %s", err.Error())
	}
	if err = writeTemplate(rootMetaDir+"license.txt", licenseTxt, pkg); err != nil {
		log.Fatalf("Could not generate license.txt for root %s", err.Error())
	}

	for platform, binary := range platforms {
		srcDir := binDir + "/" + platform + "/"
		if _, err := os.Stat(srcDir); err != nil {
			log.Printf("Platform %s does not exist", platform)
			continue
		}

		// Create meta dir
		buffer.Truncate(reset)
		buffer.WriteString(packageName + ".")
		buffer.WriteString(platform)
		buffer.WriteString("/meta/")
		metaDir := buffer.String()
		newDir(metaDir)

		// Create data dir
		buffer.Truncate(reset)
		buffer.WriteString(packageName + ".")
		buffer.WriteString(platform)
		buffer.WriteString("/data/qpm/")
		dataDir := buffer.String()
		newDir(dataDir)

		if err := copyBinary(srcDir+binary, dataDir+binary); err != nil {
			log.Fatalf("Cannot copy binary to %s: %s", dataDir, err.Error())
		}

		pkg.Platform = platform
		if err := writeTemplate(metaDir+"package.xml", platformPackageXml, pkg); err != nil {
			log.Fatalf("Could not generate package.xml for %s: %s", platform, err.Error())
		}
	}
}
