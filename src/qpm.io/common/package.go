// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package common

import (
	//	"crypto/x509/pkix"
	//	"debug/elf"
	"fmt"
	"os"
	msg "qpm.io/common/messages"
	json "github.com/golang/protobuf/jsonpb"
	"qpm.io/qpm/core"
	"regexp"
	"strings"
	"path/filepath"
)

const (
	ERR_REQUIRED_FIELD  = "%s is a required field"
	ERR_FORMATTED_FIELD = "%s requires a specific format"
)

var (
	regexPackageName = regexp.MustCompile("^[a-zA-Z]{2,}\\.[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]?(\\.[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]?)+$")
	regexVersion     = regexp.MustCompile("[0-9].[0-9].[0-9]*")
	regexAuthorName  = regexp.MustCompile("^[\\p{L}\\s'.-]+$")
	regexAuthorEmail = regexp.MustCompile(".+@.+\\..+")
	regexGitSha1     = regexp.MustCompile("^[a-fA-F0-9]{8,}$")
)

func dotSlash(dots string) string {
	return strings.Replace(dots, ".", "/", -1)
}

func dotUnderscore(dots string) string {
	return strings.Replace(dots, ".", "_", -1)
}

// Takes a name@version string and return just name
func packageName(release string) string {
	return strings.Split(release, "@")[0]
}

type DependencyList map[string]string

// Creates a new DependencyList (which is really a map) which takes a list of package
// names of the form "package@version" and produces a map of "package => "version".
// Passing in multiple versions of the same package will overwrite with the last one.
func NewDependencyList(packages []string) DependencyList {
	deps := DependencyList{}
	for _, dep := range packages {
		parts := strings.Split(dep, "@")
		pName := strings.ToLower(parts[0])

		if len(parts) > 1 {
			deps[pName] = strings.ToLower(parts[1])
		} else {
			deps[pName] = ""
		}
	}
	return deps
}

func NewPackage() *msg.Package {
	return &msg.Package{
		Name:        "",
		Description: "",
		Version: &msg.Package_Version{
			Label: "0.0.1",
			Fingerprint: "",
		},
		Author: &msg.Package_Author{
			Name:  "",
			Email: "",
		},
		Dependencies: []string{},
		Repository: &msg.Package_Repository{
			Type: msg.RepoType_GITHUB,
			Url:  "",
		},
		License: msg.LicenseType_MIT,
	}
}

type PackageWrapper struct {
	*msg.Package
	ID int // only used on server
	FilePath string
}

func NewPackageWrapper(file string) *PackageWrapper {
	return &PackageWrapper{
		Package: &msg.Package{

		},
		FilePath: file,
	}
}

func LoadPackage(path string) (*PackageWrapper, error) {
	var err error
	pw := &PackageWrapper{Package: &msg.Package{}}

	packageFile := filepath.Join(path, core.PackageFile)

	if _, err = os.Stat(packageFile); err == nil {
		var file *os.File

		if file, err = os.Open(packageFile); err != nil {
			return pw, err
		}
		defer file.Close()

		if err = json.Unmarshal(file, pw.Package); err != nil {
			return pw, err
		}

		pw.FilePath, err = filepath.Abs(file.Name())
	}

	return pw, err
}

func LoadPackages(vendorDir string) (map[string]*PackageWrapper, error) {
	packageMap := make(map[string]*PackageWrapper)

	if _, err := os.Stat(vendorDir); err != nil {
		return packageMap, err
	}

	err := filepath.Walk(vendorDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Base(path) == core.PackageFile {
			pkg, err := LoadPackage(filepath.Dir(path))
			if err != nil {
				return err
			}
			packageMap[pkg.Name] = pkg

			// found what we're looking for so skip the rest
			return filepath.SkipDir
		}
		return nil
	})

	return packageMap, err
}

func (pw PackageWrapper) Save() error {
	var file *os.File
	var err error

	// re-create the core.PackageFile file
	file, err = os.Create(core.PackageFile)

	if err != nil {
		return err
	}
	defer file.Close()

	marshaller := &json.Marshaler{
		EnumsAsInts: false,
		Indent: "  ",
	}
	return marshaller.Marshal(file, pw.Package)
}

// Remove a package from this package's list of dependencies.
func (pw *PackageWrapper) RemoveDependency(dep *PackageWrapper) {
	for i, d := range pw.Dependencies {
		if packageName(d) == dep.Name {
			pw.Dependencies = append(pw.Dependencies[:i], pw.Dependencies[i+1:]...)
			return
		}
	}
}

func (pw PackageWrapper) ParseDependencies() DependencyList {
	return NewDependencyList(pw.Dependencies)
}

func (pw PackageWrapper) Validate() error {
	if pw.Name == "" {
		return fmt.Errorf(ERR_REQUIRED_FIELD, "name")
	} else {
		// Validate name
		if !regexPackageName.MatchString(pw.Name) {
			return fmt.Errorf(ERR_FORMATTED_FIELD, "name")
		}
	}
	if pw.Version == nil {
		return fmt.Errorf(ERR_REQUIRED_FIELD, "version")
	} else {
		// Validate version label
		if !regexVersion.MatchString(pw.Version.Label) {
			return fmt.Errorf(ERR_FORMATTED_FIELD, "version label")
		}
		// Validate version revision
		if pw.Version.Revision == "" {
			return fmt.Errorf(ERR_REQUIRED_FIELD, "version revision")
		}
	}
	if pw.Author == nil {
		return fmt.Errorf(ERR_REQUIRED_FIELD, "author")
	} else {
		// Validate author name
		if !regexAuthorName.MatchString(pw.Author.Name) {
			return fmt.Errorf(ERR_FORMATTED_FIELD, "author name")
		}
		//Validate author email
		if !regexAuthorEmail.MatchString(pw.Author.Email) {
			return fmt.Errorf(ERR_FORMATTED_FIELD, "author email")
		}
	}

	return nil
}

func (pw PackageWrapper) RootDir() string {
	return filepath.Dir(pw.FilePath)
}

func (pw PackageWrapper) PriFile() string {
	if pw.PriFilename != "" {
		return pw.PriFilename
	}
	return dotUnderscore(pw.Package.Name) + ".pri"
}

func (pw PackageWrapper) QrcFile() string {
	return dotUnderscore(pw.Package.Name) + ".qrc"
}

func (pw PackageWrapper) QrcPrefix() string {
	return dotSlash(pw.Package.Name)
}

func (pw PackageWrapper) GetDependencySignature() string {
	return strings.Join([]string{pw.Name, pw.Version.Label}, "@")
}
