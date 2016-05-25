// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package vcs

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"qpm.io/common"
	msg "qpm.io/common/messages"
)

const (
	GitHubURL = "https://api.github.com/repos"
	Tarball   = "tarball"
	TarSuffix = ".tar.gz"
)

type GitHub struct {
}

func NewGitHub() *GitHub {
	return &GitHub{}
}

func (g *GitHub) Install(repository *msg.Package_Repository, version *msg.Package_Version, destination string) (*common.PackageWrapper, error) {

	destinationTokens := strings.Split(destination, "/")
	destinationSuffix := destinationTokens[len(destinationTokens)-1]
	fileDestination := strings.TrimSuffix(destination, destinationSuffix)

	os.RemoveAll(destination)

	if err := os.MkdirAll(fileDestination, 0755); err != nil {
		return nil, err
	}

	fileName, err := g.download(repository, version, fileDestination)
	if err != nil {
		return nil, err
	}

	pkg, err := g.extract(fileName, fileDestination, destinationSuffix)
	if err != nil {
		return nil, err
	}

	return pkg, os.Remove(fileName)
}

func (g *GitHub) download(repository *msg.Package_Repository, version *msg.Package_Version, destination string) (fileName string, err error) {

	repo := repository.Url
	if strings.HasPrefix(repo, "git@github.com:") {
		repo = strings.TrimPrefix(repo, "git@github.com:")
		repo = strings.TrimSuffix(repo, ".git")
	} else if strings.HasPrefix(repo, "https://github.com/") {
		repo = strings.TrimPrefix(repo, "https://github.com/")
		repo = strings.TrimSuffix(repo, ".git")
	} else {
		return "", fmt.Errorf("This does not seem to be a GitHub repository.")
	}

	url := GitHubURL + "/" + repo + "/" + Tarball + "/" + version.Revision
	urlTokens := strings.Split(url, "/")

	fileName = destination + string(filepath.Separator) + urlTokens[len(urlTokens)-1] + TarSuffix // FIXME: we assume it's a tarball

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

func (g *GitHub) extract(fileName string, destination string, suffix string) (*common.PackageWrapper, error) {

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
			tokens := strings.Split(header.Name, "/")
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

		path := filepath.Join(destination, suffix)

		if err = os.Rename(src, path); err != nil {
			return pkg, err
		}

		return pkg, err
	}

	return nil, nil
}
