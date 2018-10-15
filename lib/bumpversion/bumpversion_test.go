package bumpversion

import (
	"fmt"
	"reflect"
	"testing"

	version "github.com/hashicorp/go-version"
)

var ParseTests = []struct {
	S        string
	Versions []*version.Version
}{
	{"", nil},
	{"test", nil},
	{"test\ntest", nil},
	{
		"v1.0.0",
		[]*version.Version{
			version.Must(version.NewVersion("v1.0.0")),
		},
	},
	{
		`
		v1.0.0
		v0.0.1
		v0.1.0
		`,
		[]*version.Version{
			version.Must(version.NewVersion("v0.0.1")),
			version.Must(version.NewVersion("v0.1.0")),
			version.Must(version.NewVersion("v1.0.0")),
		},
	},
	{
		// git ls-remote --tags
		`
		a8545a6c6210ba1fdf6081bf2292c8d01cb6af9f	refs/tags/v1.0.0
		4a3ba460ffcb84477615ed0bba9f30d040fd10f1	refs/tags/v0.0.1
		ef3fe16be2ea20f7dcb77afba83bee2cb02be723	refs/tags/v0.1.0
		`,
		[]*version.Version{
			version.Must(version.NewVersion("v0.0.1")),
			version.Must(version.NewVersion("v0.1.0")),
			version.Must(version.NewVersion("v1.0.0")),
		},
	},
}

func TestParse(t *testing.T) {
	for _, tt := range ParseTests {
		v := Parse(tt.S)
		if !reflect.DeepEqual(v, tt.Versions) {
			t.Errorf("%s: want %v, got %v", tt.S, tt.Versions, v)
		}
	}
}

var BumpTests = []struct {
	From    string
	Segment Segment
	To      string
}{
	// Easy.
	{"v0.0.0", Major, "v1.0.0"},
	{"v0.0.0", Minor, "v0.1.0"},
	{"v0.0.0", Patch, "v0.0.1"},
	// Advanced.
	{"v0.1.1", Major, "v1.0.0"},
	{"v0.0.1", Minor, "v0.1.0"},
}

func TestBump(t *testing.T) {
	for _, tt := range BumpTests {
		t.Run(fmt.Sprintf("%s - %s", tt.From, tt.To), func(t *testing.T) {
			v := version.Must(version.NewVersion(tt.From))
			r := version.Must(version.NewVersion(tt.To))
			if u := Bump(v, tt.Segment); !u.Equal(r) {
				t.Error(v)
			}
		})
	}
}
