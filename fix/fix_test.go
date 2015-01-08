package fix

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghthor/journal/fix/case_0_static"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/git/gittest"
	"github.com/ghthor/journal/idea"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeAFixableJournal(c gospec.Context) {
	tmpDir := func(prefix string) (directory string, cleanUp func()) {
		directory, err := ioutil.TempDir("", prefix+"_")
		c.Assume(err, IsNil)

		cleanUp = func() {
			c.Assume(os.RemoveAll(directory), IsNil)
		}

		return
	}

	baseDir, cleanUp := tmpDir("journal_fix")
	defer cleanUp()

	c.Specify("a fixed journal is a", func() {
		// TODO this is a hack to create a fixed journal repo
		d, _, err := case_0_static.NewIn(baseDir)
		c.Assume(err, IsNil)

		_, err = Fix(d)
		c.Assume(err, IsNil)

		c.Specify("directory", func() {
			fi, err := os.Stat(d)
			c.Assume(err, IsNil)
			c.Expect(fi.IsDir(), IsTrue)

			c.Specify("inside a git repository", func() {
				c.Expect(d, gittest.IsAGitRepository)
			})

			c.Specify("containing", func() {
				c.Specify("an entry directory", func() {
					entriesPath := filepath.Join(d, "entry")
					fi, err := os.Stat(entriesPath)
					c.Assume(err, IsNil)

					c.Expect(fi.IsDir(), IsTrue)

					c.Specify("with all entries using the current entry format", func() {
						entriesDir, err := os.Open(entriesPath)
						c.Assume(err, IsNil)

						infos, err := entriesDir.Readdir(0)
						c.Assume(err, IsNil)

						for _, info := range infos {
							//TODO use a public interface to check this
							entry, err := newEntryFromFile(filepath.Join(entriesPath, info.Name()))
							c.Assume(err, IsNil)

							c.Expect(entry.needsFixed(), IsFalse)
						}
					})
				})

				c.Specify("an idea directory store", func() {
					_, err := idea.NewDirectoryStore(filepath.Join(d, "idea"))
					c.Assume(err, IsNil)

					//TODO Make sure ideas exist
					//TODO Make sure ideas are edittable
				})
			})
		})
	})

	c.Specify("a fixable journal is a", func() {
		d, _, err := case_0_static.NewIn(baseDir)
		c.Assume(err, IsNil)

		c.Specify("directory", func() {
			c.Specify("inside a git repository", func() {
				c.Assume(d, gittest.IsAGitRepository)
				c.Assume(git.IsClean(d), IsNil)

				// Case 0
				c.Specify("that contains entries", func() {
					needsFixed, err := NeedsFixed(d)
					c.Expect(needsFixed, IsTrue)

					_, err = Fix(d)
					c.Expect(err, IsNil)

					needsFixed, err = NeedsFixed(d)
					c.Expect(needsFixed, IsFalse)
				})
			})
		})
	})
}
