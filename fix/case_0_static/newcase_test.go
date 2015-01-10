package case_0_static_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghthor/journal/fix/case_0_static"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/git/gittest"

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
				c.Expect(len(entries), Equals, 7)

				case_0_dir, err := os.Open(d)
				c.Assume(err, IsNil)

				infos, err := case_0_dir.Readdir(0)
				c.Assume(err, IsNil)

				entryInfos := make([]os.FileInfo, 0, len(infos)-1)

				for _, info := range infos {
					if info.IsDir() {
						c.Expect(info.Name(), Equals, ".git")
					} else {
						entryInfos = append(entryInfos, info)
					}
				}

				c.Expect(len(entryInfos), Equals, len(entries))
			})

			c.Specify("as a git repository", func() {
				c.Expect(d, gittest.IsAGitRepository)
				c.Expect(git.IsClean(d), IsNil)

				c.Specify("and contains committed entry", func() {
					for i := 0; i < len(entries); i++ {
						entryFilename := entries[i]

						c.Specify(entryFilename, func() {
							// Check that the files were commited in the correct order
							o, err := git.Command(d, "show", "--name-only", "--pretty=format:",
								fmt.Sprintf("HEAD%s", strings.Repeat("^", len(entries)-1-i))).Output()
							c.Assume(err, IsNil)
							c.Expect(strings.TrimSpace(string(o)), Equals, entryFilename)
						})
					}

					// Verify the git tree hash is the same
					o, err := git.Command(d, "show", "-s", "--pretty=format:%T").Output()
					c.Assume(err, IsNil)
					c.Expect(string(o), Equals, "1731d5a3e0e5f6efacfee953262fe8bc82cc9a2e")
				})
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
