// Copyright 2016 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package vcs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"qpm.io/common"
	msg "qpm.io/common/messages"
)

type Git struct {
}

func NewGit() *Git {
	return &Git{}
}

func (g *Git) Install(repository *msg.Package_Repository, version *msg.Package_Version, destination string) (*common.PackageWrapper, error) {

	repo := strings.TrimSuffix(repository.Url, ".git")
	tokens := strings.Split(repo, "/")
	path := destination + string(filepath.Separator) + tokens[len(tokens)-2] + string(filepath.Separator) + tokens[len(tokens)-1]

	err := os.RemoveAll(path)
	if err != nil {
		return nil, err
	}

	err = g.cloneRepository(repository.Url, path)
	if err != nil {
		return nil, err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	err = os.Chdir(path)
	if err != nil {
		return nil, err
	}

	err = g.checkoutRevision(version.Revision)
	if err != nil {
		return nil, err
	}

	err = os.Chdir(pwd)
	if err != nil {
		return nil, err
	}

	return common.LoadPackage(path)
}

func (g *Git) Test() error {
	_, err := exec.Command("git", "version").Output()
	if err != nil {
		return err
	}
	return nil
}

func (g *Git) cloneRepository(url string, destdir string) error {
	//log.Print("git clone ", url, " ", destdir)
	_, err := exec.Command("git", "clone", url, destdir).Output()
	if err != nil {
		return err
	}
	return nil
}

func (g *Git) checkoutRevision(revision string) error {
	//log.Print("git checkout ", revision)
	_, err := exec.Command("git", "checkout", revision).Output()
	if err != nil {
		return err
	}
	return nil
}

func (g *Git) CreateTag(name string) error {
	_, err := exec.Command("git", "tag", name).Output()
	if err != nil {
		return err
	}
	return nil
}

func (g *Git) ValidateCommit(commit string) error {
	// First run 'git ls-remote' to get the HEADs of all published remotes
	lsRemote := exec.Command("git", "ls-remote")
	stdout, err := lsRemote.StdoutPipe()
	if err != nil {
		return err
	}

	if err = lsRemote.Start(); err != nil {
		return err
	}

	remotes := make(map[string]string)
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		sha1 := scanner.Text()
		if !scanner.Scan() {
			break
		}
		remote := scanner.Text()
		remotes[remote] = sha1
	}

	if err = lsRemote.Wait(); err != nil {
		return err
	}

	// Check if commit is an ancestor of a published remote
	for _, s := range remotes {
		err = exec.Command("git", "merge-base", "--is-ancestor", commit, s).Run()
		if err == nil {
			return nil
		}
	}

	if err != nil {
		err = fmt.Errorf("Commit %s has not been pushed yet.", commit[:8])
	}

	return err
}

func (g *Git) RepositoryURL() (string, error) {
	out, err := exec.Command("git", "config", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("We could not get the repository remote origin URL.")
	}
	return strings.TrimSpace(string(out)), err
}

func (g *Git) RepositoryFileList() ([]string, error) {
	var paths []string
	out, err := exec.Command("git", "ls-files").Output()
	if err != nil {
		return paths, err
	}

	// TODO: this may not work on Windows - we need to test this
	output := string(out)
	paths = strings.Split(strings.Trim(output, "\n"), "\n")

	return paths, nil
}

func (g *Git) LastCommitRevision() (string, error) {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	return strings.TrimSpace(string(out)), err
}

func (g *Git) LastCommitAuthorName() (string, error) {
	args := []string{"log", "-1", "--format=%an"}
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func (g *Git) LastCommitEmail() (string, error) {
	args := []string{"log", "-1", "--format=%ae"}
	out, err := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out)), err
}
