// Copyright 2016 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package vcs

import (
	"fmt"
	"os"

	"qpm.io/common"
	msg "qpm.io/common/messages"
)

// Installer - generic interface to functionality needed to install packages
type Installer interface {
	Install(repository *msg.Package_Repository, version *msg.Package_Version, destination string) (*common.PackageWrapper, error)
}

func CreateInstaller(repository *msg.Package_Repository) (Installer, error) {
	// TODO: add support for Mercurial
	switch repository.Type {
	case msg.RepoType_GIT:
		git := NewGit()
		if err := git.Test(); err == nil {
			return git, nil
		}
		fallthrough
	case msg.RepoType_GITHUB:
		return NewGitHub(), nil
	case msg.RepoType_MERCURIAL:
		hg := NewMercurial()
		if err := hg.Test(); err == nil {
			return hg, nil
		}
	}
	return nil, fmt.Errorf("Repository type %d is not supported", repository.Type)
}

// Publisher - generic interface to VCS functionality need to publish packages
type Publisher interface {
	Test() error
	CreateTag(name string) error
	ValidateCommit(commit string) error
	RepositoryURL() (string, error)
	LastCommitRevision() (string, error)
	LastCommitAuthorName() (string, error)
	LastCommitEmail() (string, error)
	RepositoryFileList() ([]string, error)
}

func CreatePublisher(repository *msg.Package_Repository) (Publisher, error) {
	// TODO: add support for Mercurial
	switch repository.Type {
	case msg.RepoType_GIT:
		fallthrough
	case msg.RepoType_GITHUB:
		git := NewGit()
		if err := git.Test(); err != nil {
			return nil, err
		}
		return git, nil
	case msg.RepoType_MERCURIAL:
		hg := NewMercurial()
		if err := hg.Test(); err != nil {
			return nil, err
		}
		return hg, nil
	}
	return nil, fmt.Errorf("Repository type %d is not supported", repository.Type)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func RepoType() (msg.RepoType, error) {
	if yes, _ := exists(".git"); yes {
		return msg.RepoType_GIT, nil
	}
	if yes, _ := exists(".hg"); yes {
		return msg.RepoType_MERCURIAL, nil
	}
	return msg.RepoType_AUTO, fmt.Errorf("Could not auto-detect the repository type")
}
