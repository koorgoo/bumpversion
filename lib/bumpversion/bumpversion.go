package bumpversion

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	version "github.com/hashicorp/go-version"
)

// VersionRegexpRaw is a very simple version regexp pattern.
var VersionRegexpRaw = `v?[0-9]+\.[0-9]+\.[0-9]+`

var versionRegexp = regexp.MustCompile(`(` + VersionRegexpRaw + `)`)

var zeroVersion = version.Must(version.NewVersion("v0.0.0"))

// ErrTagsDiffer is returned when remote tags differ from local ones.
var ErrTagsDiffer = errors.New("bumpversion: tags differ on remote")

// VCS defines an interface for a minimal VCS.
type VCS interface {
	// Tags returns a list of consistent tags. It may return ErrTagsDiffer.
	Tags() ([]*version.Version, error)
	// Tags add a tag reference to a current commit.
	Tag(*version.Version) error
	// Push pushes tags to remote.
	Push() error
}

// Segment is a version segment.
type Segment int

// Version segments.
const (
	Major Segment = iota
	Minor
	Patch
)

// Bumper defines an interface to bump version segment.
type Bumper interface {
	Bump(Segment) (string, error)
}

// NewBumper initializes and returns Bumper.
func NewBumper() (b Bumper, err error) {
	v, err := NewGit(GitConfig{})
	if err != nil {
		return
	}
	b = &bumper{
		vcs: v,
	}
	return
}

type bumper struct {
	vcs VCS
}

func (b *bumper) Bump(seg Segment) (ver string, err error) {
	a, err := b.vcs.Tags()
	if err != nil {
		return
	}
	if len(a) == 0 {
		a = append(a, zeroVersion)
	}
	v := Bump(a[len(a)-1], seg)
	if err = b.vcs.Tag(v); err != nil {
		return
	}
	if err = b.vcs.Push(); err != nil {
		return
	}
	ver = Format(v)
	return
}

// Bump bumps segment in version.
func Bump(v *version.Version, s Segment) *version.Version {
	a := v.Segments()
	a[s]++
	for i := s + 1; i <= Patch && i < Segment(len(a)); i++ {
		a[i] = 0
	}
	t := FormatSegments(a[0], a[1], a[2])
	return version.Must(version.NewVersion(t))
}

// Parse returns a sorted list of versions parsed from s.
func Parse(s string) (v []*version.Version) {
	for _, line := range strings.Split(s, "\n") {
		m := versionRegexp.FindStringSubmatch(line)
		if m != nil {
			t, err := version.NewVersion(m[1])
			if err != nil {
				panic(err)
			}
			v = append(v, t)
		}
	}
	sort.Sort(version.Collection(v))
	return
}

// Format formats a version.
func Format(v *version.Version) string {
	return "v" + v.String()
}

// FormatSegments formats a version from segments.
func FormatSegments(major, minor, patch int) string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

// Run runs cmd and pipes stdout/stderr on error only.
func Run(cmd *exec.Cmd) (err error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		io.Copy(os.Stdout, &stdout)
		io.Copy(os.Stderr, &stderr)
	}
	return err
}
