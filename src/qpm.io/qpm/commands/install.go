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
	"path/filepath"
	"qpm.io/common"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
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
}

var (
	vendorPri = template.Must(template.New("vendorPri").Funcs(packageFuncs).Parse(`
DEFINES += QPM_INIT\\(E\\)=\"E.addImportPath(QStringLiteral(\\\"qrc:/\\\"));\"

{{$vendirDir := .VendorDir}}
{{range $dep := .Dependencies}}
include($$PWD/{{relPriFile $vendirDir $dep}})
{{end}}
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
	response, err := i.Ctx.Client.GetDependencies(context.Background(), &msg.DependencyRequest{packageNames})
	if err != nil {
		i.Error(err)
		return err
	}

	if len(response.Dependencies) == 0 {
		i.Info("No package(s) found")
		return nil
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

	err = i.postInstall(packages)
	// FIXME: should we continue installing ?
	if err != nil {
		return err
	}

	return nil
}

func (i *InstallCommand) install(d *msg.Dependency) (*common.PackageWrapper, error) {

	url := core.GitHub + "/" + d.Repository.Url + "/" + core.Tarball

	signature := strings.Join([]string{d.Name, d.Version.Label}, "@")
	fmt.Println("Installing", signature)

	fileName, err := i.download(url, i.vendorDir)
	if err != nil {
		return nil, err
	}

	pkg, err := i.extract(fileName, i.vendorDir)
	if err != nil {
		return nil, err
	}

	return pkg, os.Remove(fileName)
}

func (i *InstallCommand) save(newDeps []*common.PackageWrapper) error {

	existingDeps := i.pkg.ParseDependencies()

	for _, d := range newDeps {
		existingVersion, exists := existingDeps[d.Name]
		if exists {
			if d.Version.Label == existingVersion {
				i.Info("The package is already a dependency : " + d.GetDependencySignature())
			} else {
				// TODO: Handle conflicts
				err := fmt.Errorf("Conflict for package %s. Version %s != %s", d.Name, existingVersion, d.Version.Label)
				i.Error(err)
				return err
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

func (i *InstallCommand) extract(fileName string, destination string) (*common.PackageWrapper, error) {

	file, err := os.Open(fileName)

	if err != nil {
		i.Error(err)
		return nil, err
	}

	defer file.Close()

	var fileReader io.ReadCloser = file

	// add a filter to handle gzipped file
	if strings.HasSuffix(fileName, ".gz") {
		if fileReader, err = gzip.NewReader(file); err != nil {
			i.Error(err)
			return nil, err
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
			return nil, err
		}

		filename := destination + "/" + header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			tokens := strings.Split(header.Name, "/")
			topDir = tokens[0]
			err = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755
			if err != nil {
				i.Error(err)
				return nil, err
			}

		case tar.TypeReg:
			writer, err := os.Create(filename)
			if err != nil {
				i.Error(err)
				return nil, err
			}
			io.Copy(writer, tarBallReader)
			err = os.Chmod(filename, os.FileMode(header.Mode))
			if err != nil {
				i.Error(err)
				return nil, err
			}
			writer.Close()

		case tar.TypeXGlobalHeader:
			// Ignore this

		default:
			//i.Info("Unable to extract type : %c in file %s\n", header.Typeflag, filename)
		}
	}

	if topDir != "" {

		src := filepath.Join(destination, topDir)

		pkg, err := common.LoadPackage(src)
		if err != nil {
			i.Error(err)
			return pkg, err
		}

		path := filepath.Join(destination, pkg.QrcPrefix())

		if err := os.MkdirAll(path, 0755); err != nil {
			i.Error(err)
			return pkg, err
		}

		os.RemoveAll(path)

		if err = os.Rename(src, path); err != nil {
			i.Error(err)
			return pkg, err
		}

		// Reload it from the new location
		pkg, err = common.LoadPackage(path)
		if err != nil {
			i.Error(err)
		}

		return pkg, err
	}

	return nil, nil
}

func (i *InstallCommand) postInstall(dependencies []*common.PackageWrapper) error {
	if err := GenerateVendorPri(i.vendorDir, i.pkg, dependencies); err != nil {
		i.Error(err)
		return err
	}
}

// Generates a vendor.pri inside vendorDir using the information contained in the package file
// and the dependencies
func GenerateVendorPri(vendorDir string, pkg *common.PackageWrapper, deps []*common.PackageWrapper) error {

	vendorPriFile := filepath.Join(vendorDir, core.Vendor+".pri")

	var file *os.File
	var err error

	// re-create the .pri file
	file, err = os.Create(vendorPriFile)

	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		VendorDir    string
		Package      *common.PackageWrapper
		Dependencies []*common.PackageWrapper
	}{
		vendorDir,
		pkg,
		deps,
	}

	if err := vendorPri.Execute(file, data); err != nil {
		return err
	}

	return nil
}
