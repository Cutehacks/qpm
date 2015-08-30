// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"os"
	"qpm.io/common"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
	"strings"
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
	pkg common.PackageWrapper
	fs  *flag.FlagSet
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
}

func (i *InstallCommand) Run() error {

	packageName := i.fs.Arg(0)

	err := i.pkg.Load()
	if err != nil {
		i.Error(err)
		return err
	}

	var packageNames []string
	if packageName == "" {
		packageNames = i.pkg.Dependencies
	} else {
		packageNames = []string{packageName}
	}

	// Get list of dependencies from the server
	response, err := i.Ctx.Client.GetDependencies(context.Background(), &msg.DependencyRequest{packageNames})
	if err != nil {
		i.Error(err)
		return err
	}

	if len(response.Dependencies) == 0 {
		i.Info("No package(s) found")
		return nil
	}

	// Download and extract the packages
	for _, d := range response.Dependencies {
		err = i.install(d)
		// FIXME: should we continue installing ?
		if err != nil {
			return err
		}
	}

	// Save the dependencies in package.json
	err = i.save(response.Dependencies)
	// FIXME: should we continue installing ?
	if err != nil {
		return err
	}

	err = i.pkg.UpdatePri(response.Dependencies)
	// FIXME: should we continue installing ?
	if err != nil {
		return err
	}

	return nil
}

func (i *InstallCommand) install(d *msg.Dependency) error {

	url := core.GitHub + "/" + d.Repository.Url + "/" + core.Tarball

	signature := i.pkg.GetDependencySignature(d)
	fmt.Println("Installing", signature)

	fileName, err := i.download(url, core.Vendor)
	if err != nil {
		return err
	}

	err = i.extract(fileName, core.Vendor, d.Name)
	if err != nil {
		return err
	}

	return os.Remove(fileName)
}

func (i *InstallCommand) save(dependencies []*msg.Dependency) error {

	// FIXME: inefficient
	var newDependencies []string
	for _, d := range dependencies {
		exists := false
		signature := i.pkg.GetDependencySignature(d)
		for _, dependency := range i.pkg.Dependencies {
			if dependency == signature {
				exists = true
				i.Info("The package is already a dependency : " + signature)
				break
			}
		}
		if !exists {
			newDependencies = append(newDependencies, signature)
		}
	}
	i.pkg.Dependencies = append(i.pkg.Dependencies, newDependencies...)

	err := i.pkg.Save()
	if err != nil {
		i.Error(err)
		return err
	}

	return nil
}

func (i *InstallCommand) download(url string, destination string) (fileName string, err error) {

	tokens := strings.Split(url, "/")
	fileName = destination + "/" + tokens[len(tokens)-2] + core.TarSuffix // FIXME: we assume it's a tarball

	var output *os.File
	output, err = os.Create(fileName)
	if err != nil {
		// TODO: check file existence first with os.IsExist(err)
		i.Error(err)
		return
	}
	defer output.Close()

	var response *http.Response
	response, err = http.Get(url)
	if err != nil {
		i.Error(err)
		return
	}
	defer response.Body.Close()

	//proxy := &ProgressProxyReader{ Reader: response.Body, length: response.ContentLength }
	//var written int64
	_, err = io.Copy(output, response.Body)
	if err != nil {
		i.Error(err)
		return
	}

	return
}

func (i *InstallCommand) extract(fileName string, destination string, name string) (error) {

	file, err := os.Open(fileName)

	if err != nil {
		i.Error(err)
		return err
	}

	defer file.Close()

	var fileReader io.ReadCloser = file

	// add a filter to handle gzipped file
	if strings.HasSuffix(fileName, ".gz") {
		if fileReader, err = gzip.NewReader(file); err != nil {
			i.Error(err)
			return err
		}
		defer fileReader.Close()
	}

	tarBallReader := tar.NewReader(fileReader)
	var topDir string

	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			i.Error(err)
			return err
		}

		filename := destination + "/" + header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			tokens := strings.Split(header.Name, "/")
			topDir = tokens[0]
			err = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755
			if err != nil {
				i.Error(err)
				return err
			}

		case tar.TypeReg:
			writer, err := os.Create(filename)
			if err != nil {
				i.Error(err)
				return err
			}
			io.Copy(writer, tarBallReader)
			err = os.Chmod(filename, os.FileMode(header.Mode))
			if err != nil {
				i.Error(err)
				return err
			}
			writer.Close()

		case tar.TypeXGlobalHeader:
			// Ignore this

		default:
			//i.Info("Unable to extract type : %c in file %s\n", header.Typeflag, filename)
		}
	}

	if topDir != "" {

		path := destination + "/" + strings.Replace(name, ".", "/", -1)

		if err := os.MkdirAll(path, 0755); err != nil {
			i.Error(err)
			return err
		}

		os.RemoveAll(path)
		err := os.Rename(destination+"/"+topDir, path)
		if err != nil {
			i.Error(err)
		}
	}

	return nil
}
