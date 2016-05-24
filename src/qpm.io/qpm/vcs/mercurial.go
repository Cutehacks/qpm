// Copyright 2016 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package vcs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"qpm.io/common"
	msg "qpm.io/common/messages"
)

type Mercurial struct {
}

func NewMercurial() *Mercurial {
	return &Mercurial{}
}

func (m *Mercurial) Install(repository *msg.Package_Repository, version *msg.Package_Version, destination string) (*common.PackageWrapper, error) {
	repo := strings.TrimSuffix(repository.Url, "/hg")
	tokens := strings.Split(repo, "/")
	path := destination + string(filepath.Separator) + tokens[len(tokens)-2] + string(filepath.Separator) + tokens[len(tokens)-1]

	err := os.RemoveAll(path)
	if err != nil {
		return nil, err
	}

	err = m.cloneRepository(repository.Url, version.Revision, path)
	if err != nil {
		return nil, err
	}

	return common.LoadPackage(path)
}

func (m *Mercurial) Test() error {
	_, err := exec.Command("hg", "version").Output()
	if err != nil {
		return err
	}
	return nil
}

func (m *Mercurial) cloneRepository(url string, revision string, destdir string) error {
	_, err := exec.Command("hg", "clone", "-r", revision, url, destdir).Output()
	if err != nil {
		return err
	}
	return nil
}

func (m *Mercurial) CreateTag(name string) error {
	_, err := exec.Command("hg", "tag", name).Output()
	if err != nil {
		return err
	}
	return nil
}

func (m *Mercurial) ValidateCommit(commit string) error {
	url, err := m.RepositoryURL()
	if err != nil {
		return err
	}
	_, err = exec.Command("hg", "identify", url, "-r", commit).Output()
	return err
}

func (m *Mercurial) RepositoryURL() (string, error) {
	out, err := exec.Command("hg", "paths", "default").Output()
	if err != nil {
		return "", fmt.Errorf("We could not get the repository default path URL.")
	}
	return strings.TrimSpace(string(out)), err
}

func (m *Mercurial) RepositoryFileList() ([]string, error) {
	var paths []string
	out, err := exec.Command("hg", "locate").Output()
	if err != nil {
		return paths, err
	}

	// TODO: this may not work on Windows - we need to test this
	output := string(out)
	paths = strings.Split(strings.Trim(output, "\n"), "\n")

	return paths, nil
}

func (m *Mercurial) LastCommitRevision() (string, error) {
	out, err := exec.Command("hg", "log", "--template", "{node}", "--limit", "1").Output()
	return strings.TrimSpace(string(out)), err
}

func (m *Mercurial) LastCommitAuthorName() (string, error) {
	out, err := exec.Command("hg", "log", "--template", "{author|person}", "--limit", "1").Output()
	return strings.TrimSpace(string(out)), err
}

func (m *Mercurial) LastCommitEmail() (string, error) {
	out, err := exec.Command("hg", "log", "--template", "{author|email}", "--limit", "1").Output()
	return strings.TrimSpace(string(out)), err
}
