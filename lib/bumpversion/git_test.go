package bumpversion

import (
	"os"
	"path"
	"testing"
)

func repoRoot() string {
	s, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return path.Dir(path.Dir(s))
}

func TestNewGit(t *testing.T) {
	v, err := NewGit(GitConfig{
		C: repoRoot(),
	})
	if err != nil {
		t.Fatal(v)
	}
	_, err = v.Tags()
	if err != nil {
		t.Fatal(err)
	}
}
