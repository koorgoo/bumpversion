package bumpversion

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

	version "github.com/hashicorp/go-version"
)

// ErrNoGit is returned when git is not in PATH.
var ErrNoGit = errors.New("bumpversion: git not found")

// GitConfig configures git.
type GitConfig struct {
	// C overrides a path to start git in.
	// It works like `-C` parameter (see `git help git`).
	// If empty, current working directory is used.
	C string
	// command is a exec.Cmd building function. Used in tests.
	command commandFunc
}

type commandFunc func(name string, arg ...string) *exec.Cmd

// NewGit returns VCS for git using cwd.
func NewGit(c GitConfig) (v VCS, err error) {
	if c.C == "" {
		c.C, err = os.Getwd()
		if err != nil {
			return
		}
	}
	if c.command == nil {
		c.command = exec.Command
	}
	g := git{
		C:       c.C,
		Remote:  "origin",
		Command: c.command,
	}
	err = g.check()
	if err != nil {
		return
	}
	v = &g
	return
}

type git struct {
	C       string
	Remote  string
	Command commandFunc
}

// check checks whether g.root is a repository.
func (g *git) check() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return ErrNoGit
	}
	err = g.Command("git", "status").Run()
	if err != nil {
		return fmt.Errorf("bumpversion: git: not a repository: %s", g.C)
	}
	err = g.Command("git", "remote", "get-url", g.Remote).Run()
	if err != nil {
		return fmt.Errorf("bumpversion: git: no such remote: %s", g.Remote)
	}
	return nil
}

func (g *git) Tags() (v []*version.Version, err error) {
	var buf bytes.Buffer
	cmd := g.Command("git", "tag")
	cmd.Stdout = &buf
	if err = cmd.Run(); err != nil {
		return
	}
	local := Parse(buf.String())

	buf.Reset()
	cmd = g.Command("git", "ls-remote", "--tags", g.Remote)
	cmd.Stdout = &buf
	if err = cmd.Run(); err != nil {
		return
	}
	remote := Parse(buf.String())

	if !equalCollections(local, remote) {
		return nil, ErrTagsDiffer
	}
	return local, nil
}

func (g *git) Tag(v *version.Version) error {
	cmd := g.Command("git", "tag", Format(v))
	return Run(cmd)
}

func (g *git) Push() error {
	cmd := g.Command("git", "push", "--tags", g.Remote)
	return Run(cmd)
}

func equalCollections(a, b version.Collection) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}
