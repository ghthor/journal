package case_0_static_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghthor/journal/fix/case_0_static"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeNewCase0)

	gospec.MainGoTest(r, t)
}

func DescribeNewCase0(c gospec.Context) {
	tmpDir := func(prefix string) (directory string, cleanUp func()) {
		directory, err := ioutil.TempDir("", prefix+"_")
		c.Assume(err, IsNil)

		cleanUp = func() {
			c.Assume(os.RemoveAll(directory), IsNil)
		}

		return
	}

	c.Specify("a new case 0 directory is created", func() {
		baseDirectory, cleanUp := tmpDir("new_case_0")
		defer cleanUp()

		d, entries, err := case_0_static.NewIn(baseDirectory)
		c.Assume(err, IsNil)

		c.Specify("with a case_0/ directory", func() {
			fi, err := os.Stat(filepath.Join(baseDirectory, "case_0/"))
			c.Assume(err, IsNil)

			c.Expect(fi.IsDir(), IsTrue)

			c.Specify("containing some entries", func() {
				c.Expect(len(entries), Equals, 6)

				case_0_dir, err := os.Open(d)
				c.Assume(err, IsNil)

				entryInfos, err := case_0_dir.Readdir(0)
				c.Assume(err, IsNil)

				c.Expect(len(entryInfos), Equals, len(entries))
			})

			c.Specify("as a git repository", func() {
				c.Specify("and contains committed entry", func() {
				})
			})

			c.Specify("that can be fixed", func() {
			})
		})

		c.Specify("with a case_0_fix_reflog/ directory", func() {
			fi, err := os.Stat(filepath.Join(baseDirectory, "case_0_fix_reflog/"))
			c.Assume(err, IsNil)

			c.Expect(fi.IsDir(), IsTrue)

			c.Specify("containing the reflog of git commits that will fix the repository", func() {
				reflogDir, err := os.Open(filepath.Join(baseDirectory, "case_0_fix_reflog"))
				c.Assume(err, IsNil)

				reflogInfos, err := reflogDir.Readdir(0)
				c.Assume(err, IsNil)

				c.Expect(len(reflogInfos), Equals, 14)
			})
		})

		c.Specify("with a case_0.json journal configuration file", func() {
			fi, err := os.Stat(filepath.Join(baseDirectory, "case_0.json"))
			c.Assume(err, IsNil)
			c.Expect(fi.IsDir(), IsFalse)
		})
	})
}
