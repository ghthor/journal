package init_test

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/entry"
	"github.com/ghthor/journal/git/gittest"
	"github.com/ghthor/journal/idea"
	jinit "github.com/ghthor/journal/init"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func DescribeInit(c gospec.Context) {
	c.Specify("a journal", func() {
		tmpJournal := func() (directory string, cleanUp func()) {
			directory, err := ioutil.TempDir("", "journal_init_")
			c.Assume(err, IsNil)

			c.Assume(jinit.Journal(directory), IsNil)

			cleanUp = func() {
				c.Assume(os.RemoveAll(directory), IsNil)
			}

			return
		}

		c.Specify("that is initialized", func() {
			jd, cleanUp := tmpJournal()
			defer cleanUp()

			c.Assume(jinit.HasBeenInitialized(jd), IsTrue)

			c.Specify("is a git repository", func() {
				c.Expect(jd, gittest.IsAGitRepository)

				c.Specify("that contains", func() {
					c.Specify("an entry directory", func() {
						info, err := os.Stat(filepath.Join(jd, "entry"))
						c.Assume(err, IsNil)
						c.Expect(info.IsDir(), IsTrue)

						c.Specify("that can have entries", func() {
							c.Expect(jinit.HasBeenInitialized(jd), IsTrue)

							ne := entry.New(filepath.Join(jd, "entry/"))
							oe, err := ne.Open(time.Now(), nil)
							c.Assume(err, IsNil)

							_, err = oe.Close(time.Now())
							c.Assume(err, IsNil)

							c.Expect(jinit.HasBeenInitialized(jd), IsTrue)
						})
					})

					c.Specify("an idea directory store", func() {
						ids, err := idea.NewDirectoryStore(filepath.Join(jd, "idea/"))
						c.Assume(err, IsNil)

						c.Specify("that can have ideas", func() {
							c.Expect(jinit.HasBeenInitialized(jd), IsTrue)

							ids.SaveIdea(&idea.Idea{
								Name: "An Idea",
								Body: "A Body\n",
							})

							c.Expect(jinit.HasBeenInitialized(jd), IsTrue)
						})
					})
				})
			})
		})

		c.Specify("that can be initialized", func() {
			c.Specify("is an empty directory", func() {
				c.Specify("inside a git repository", func() {
				})

				c.Specify("NOT inside a git repository", func() {
				})
			})

			c.Specify("is a directory that doesn't exist", func() {
			})
		})

		c.Specify("that can NOT be initialized", func() {
			c.Specify("is a directory", func() {
				c.Specify("that already contains", func() {
					c.Specify("an entry directory", func() {
					})

					c.Specify("an idea directory store", func() {
					})
				})
			})
		})
	})
}
