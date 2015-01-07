package fix

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	entryPkg "github.com/ghthor/journal/entry"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeEntriesCollector(c gospec.Context) {
	c.Specify("collects all the entries", func() {
		tmpDir := func(prefix string) (directory string, cleanUp func(), err error) {
			directory, err = ioutil.TempDir("", prefix+"_")

			cleanUp = func() {
				c.Assume(os.RemoveAll(directory), IsNil)
			}

			return
		}

		c.Specify("in a directory containing", func() {
			c.Specify("only entries", func() {
				d, cleanUp, err := tmpDir("only_entries")
				c.Assume(err, IsNil)
				defer cleanUp()

				entryFilenames := []string{
					time.Date(2015, time.January, 1, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
					time.Date(2015, time.January, 2, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
					time.Date(2015, time.January, 3, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
					time.Date(2015, time.January, 4, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
				}

				for _, filename := range entryFilenames {
					_, err := os.Create(filepath.Join(d, filename))
					c.Assume(err, IsNil)
				}

				entries, err := entriesIn(d)
				c.Assume(err, IsNil)

				c.Expect(entries, ContainsExactly, entryFilenames)
			})

			c.Specify("entries and non entries", func() {
				d, cleanUp, err := tmpDir("entries_and_other_files")
				c.Assume(err, IsNil)
				defer cleanUp()

				entryFilenames := []string{
					time.Date(2015, time.January, 1, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
					time.Date(2015, time.January, 2, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
					time.Date(2015, time.January, 3, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
					time.Date(2015, time.January, 4, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
				}

				otherFilenames := []string{
					"notanentry",
					"somedirectory/somefile",
				}

				allFiles := append(entryFilenames, otherFilenames...)

				c.Assume(os.Mkdir(filepath.Join(d, "somedirectory"), 0777), IsNil)
				for _, filename := range allFiles {
					_, err := os.Create(filepath.Join(d, filename))
					c.Assume(err, IsNil)
				}

				entries, err := entriesIn(d)
				c.Assume(err, IsNil)

				c.Expect(entries, ContainsExactly, entryFilenames)
			})

			c.Specify("no entries", func() {
				d, cleanUp, err := tmpDir("no_entries")
				c.Assume(err, IsNil)
				defer cleanUp()

				otherFilenames := []string{
					"notanentry",
					"somedirectory/somefile",
				}

				c.Assume(os.Mkdir(filepath.Join(d, "somedirectory"), 0777), IsNil)
				for _, filename := range otherFilenames {
					_, err := os.Create(filepath.Join(d, filename))
					c.Assume(err, IsNil)
				}

				entries, err := entriesIn(d)
				c.Assume(err, IsNil)

				c.Expect(len(entries), Equals, 0)
			})
		})
	})
}
