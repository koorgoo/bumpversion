package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/koorgoo/bumpversion/lib/bumpversion"
)

// Errors.
var (
	ErrNoVersion  = errors.New("bumpversion: no version")
	ErrOneVersion = errors.New("bumpversion: choose one version")
)

func main() {
	var c Config

	flag.BoolVar(&c.Major, "major", false, "bump major version")
	flag.BoolVar(&c.Minor, "minor", false, "bump minor version")
	flag.BoolVar(&c.Patch, "patch", false, "bump patch version")
	flag.Parse()

	seg, err := c.Segment()
	exitOnError(err)

	b, err := bumpversion.NewBumper()
	exitOnError(err)

	s, err := b.Bump(seg)
	exitOnError(err)

	fmt.Println(s)
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Config helps to parse flags.
type Config struct {
	Major, Minor, Patch bool
}

// Segment returns a segment to bump.
func (c *Config) Segment() (v bumpversion.Segment, err error) {
	n := bti(c.Major) + bti(c.Minor) + bti(c.Patch)
	if n == 0 {
		err = ErrNoVersion
		return
	}
	if n != 1 {
		err = ErrOneVersion
		return
	}
	switch {
	case c.Major:
		v = bumpversion.Major
	case c.Minor:
		v = bumpversion.Minor
	case c.Patch:
		v = bumpversion.Patch
	}
	return
}

func bti(b bool) int {
	if b {
		return 1
	}
	return 0
}
