// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package vcs

import (
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
	"path/filepath"
	"io"
	"os"
	"os/exec"
	"compress/gzip"
	"archive/tar"
	"qpm.io/common"
	msg "qpm.io/common/messages"
)

const (
	GitHubURL     = "https://api.github.com/repos"
	Tarball       = "tarball"
	TarSuffix     = ".tar.gz"
)

func LastCommitSHA1() (string, error) {
	// TODO: refactor this to an interface for all VCSs
	out, err := exec.Command("git","rev-parse", "HEAD").Output()
	return strings.TrimSpace(string(out)), err
}

func LastCommitAuthorName() (string, error) {
	// TODO: refactor this to an interface for all VCSs
	args := []string{"log", "-1", "--format=%an"}
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func LastCommitEmail() (string, error) {
	// TODO: refactor this to an interface for all VCSs
	args := []string{"log", "-1", "--format=%ae"}
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func RepositoryURL() (string, error) {
	// TODO: refactor this to an interface for all VCSs and hosts
	out, err := exec.Command("git", "config", "remote.origin.url").Output()
	if err != nil {
		return "", err
	}
	// FIXME: we require the URL to be https:// and assume the GitHub API when downloading tarballs
	str := strings.TrimSpace(string(out))
	if strings.HasPrefix(str, "git@") {
		str = strings.Replace(str, ":", "/", -1)
		str = strings.Replace(str, "git@", "https://", -1)
		str = strings.TrimSuffix(str, ".git")
	}
	return str, nil
}

func RepositoryFileList() ([]string, error) {
	// TODO: refactor this to an interface for all VCSs

	var paths []string
	out, err := exec.Command("git","ls-files").Output()
	if err != nil {
		return paths, err
	}

	// TODO: this may not work on Windows - we need to test this
	output := string(out)
	paths = strings.Split(strings.Trim(output, "\n"), "\n")

	return paths, nil
}

func Install(dependency *msg.Dependency, destination string) (*common.PackageWrapper, error) {
	fileName, err := download(dependency, destination)
	if err != nil {
		return nil, err
	}

	pkg, err := extract(fileName, destination)
	if err != nil {
		return nil, err
	}

	return pkg, os.Remove(fileName)
}

func download(dependency *msg.Dependency, destination string) (fileName string, err error) {

	url := GitHubURL + "/" + dependency.Repository.Url + "/" + Tarball
	tokens := strings.Split(url, "/")
	fileName = destination + string(filepath.Separator) + tokens[len(tokens)-2] + TarSuffix // FIXME: we assume it's a tarball

	var output *os.File
	output, err = os.Create(fileName)
	if err != nil {
		// TODO: check file existence first with os.IsExist(err)
		return "", err
	}
	defer output.Close()

	var response *http.Response
	response, err = http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		errResp := make(map[string]string)
		dec := json.NewDecoder(response.Body)
		if err = dec.Decode(&errResp); err != nil {
			return "", err
		}
		errMsg, ok := errResp["message"]
		if !ok {
			errMsg = response.Status
		}
		err = fmt.Errorf("Error fetching %s: %s", url, errMsg)
		return "", err
	}

	//proxy := &ProgressProxyReader{ Reader: response.Body, length: response.ContentLength }
	//var written int64
	_, err = io.Copy(output, response.Body)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func extract(fileName string, destination string) (*common.PackageWrapper, error) {

	file, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var fileReader io.ReadCloser = file

	// add a filter to handle gzipped file
	if strings.HasSuffix(fileName, ".gz") {
		if fileReader, err = gzip.NewReader(file); err != nil {
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
			return nil, err
		}

		filename := destination + string(filepath.Separator) + header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			tokens := strings.Split(header.Name, string(filepath.Separator))
			topDir = tokens[0]
			err = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755
			if err != nil {
				return nil, err
			}

		case tar.TypeReg:
			writer, err := os.Create(filename)
			if err != nil {
				return nil, err
			}
			io.Copy(writer, tarBallReader)
			err = os.Chmod(filename, os.FileMode(header.Mode))
			if err != nil {
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
			return pkg, err
		}

		path := filepath.Join(destination, pkg.QrcPrefix())

		if err := os.MkdirAll(path, 0755); err != nil {
			return pkg, err
		}

		os.RemoveAll(path)

		if err = os.Rename(src, path); err != nil {
			return pkg, err
		}

		return pkg, err
	}

	return nil, nil
}