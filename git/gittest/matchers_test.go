package gittest

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeMatchers)

	gospec.MainGoTest(r, t)
}

func DescribeMatchers(c gospec.Context) {
	c.Specify("a directory", func() {
		d, err := ioutil.TempDir("", "gittest_")
		c.Assume(err, IsNil)

		c.Assume(d, Not(IsAGitRepository))
		c.Assume(d, Not(IsInsideAGitRepository))
		c.Assume(git.Init(d), IsNil)

		defer func() {
			c.Assume(os.RemoveAll(d), IsNil)
		}()

		c.Specify("is a git repository", func() {
			c.Expect(d, IsAGitRepository)
		})

		c.Specify("is inside a git repository", func() {
			c.Expect(d, IsInsideAGitRepository)

			sd := filepath.Join(d, "subdirectory")
			c.Assume(os.Mkdir(sd, 0755), IsNil)

			c.Expect(sd, IsInsideAGitRepository)
		})
	})
}
