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

		createFiles := func(inDirectory string, filenames []string) []string {
			for _, filename := range filenames {
				_, err := os.Create(filepath.Join(inDirectory, filename))
				c.Assume(err, IsNil)
			}

			return filenames
		}

		createSomeFiles := func(inDirectory string) (filenames []string) {
			filenames = []string{
				"notanentry",
				"jpg.bpg",
			}

			return createFiles(inDirectory, filenames)
		}

		createSomeEntries := func(inDirectory string) (entryFilenames []string) {
			entryFilenames = []string{
				time.Date(2015, time.January, 1, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
				time.Date(2015, time.January, 2, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
				time.Date(2015, time.January, 3, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
				time.Date(2015, time.January, 4, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
			}

			return createFiles(inDirectory, entryFilenames)
		}

		c.Specify("in a directory containing", func() {
			c.Specify("only entries", func() {
				d, cleanUp, err := tmpDir("only_entries")
				c.Assume(err, IsNil)
				defer cleanUp()

				entryFilenames := createSomeEntries(d)

				entries, err := entriesIn(d)
				c.Assume(err, IsNil)

				c.Expect(entries, ContainsExactly, entryFilenames)
			})

			c.Specify("entries and non entries", func() {
				d, cleanUp, err := tmpDir("entries_and_other_files")
				c.Assume(err, IsNil)
				defer cleanUp()

				entryFilenames := createSomeEntries(d)
				otherFilenames := createSomeFiles(d)

				entries, err := entriesIn(d)
				c.Assume(err, IsNil)

				c.Expect(entries, ContainsExactly, entryFilenames)
				c.Expect(entries, Not(ContainsAny), otherFilenames)
			})

			createSubDirectoriedFiles := func(baseD string, pathMap map[string][]string) {
				for d, filenames := range pathMap {
					c.Assume(len(filenames), Not(Equals), 0)
					c.Assume(os.Mkdir(filepath.Join(baseD, d), 0777), IsNil)

					for _, filename := range filenames {
						_, err := os.Create(filepath.Join(baseD, d, filename))
						c.Assume(err, IsNil)
					}
				}
			}

			c.Specify("a subdirectory", func() {
				d, cleanUp, err := tmpDir("entries_and_subdirectory")
				c.Assume(err, IsNil)
				defer cleanUp()

				subdirectoriedFiles := map[string][]string{
					"somedirectory": []string{
						"somefile",
						"anotherfile",
					},
					"image": []string{
						"jpg.bpg",
					},
				}

				entryFilenames := createSomeEntries(d)
				createSomeFiles(d)
				createSubDirectoriedFiles(d, subdirectoriedFiles)

				entries, err := entriesIn(d)
				c.Assume(err, IsNil)

				c.Expect(entries, ContainsExactly, entryFilenames)

				c.Specify("containing some entries that will be ignored", func() {
					subdirectoriedEntries := map[string][]string{
						"entry": []string{
							time.Date(2015, time.January, 5, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
							time.Date(2015, time.January, 6, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
							time.Date(2015, time.January, 7, 0, 0, 0, 0, time.Local).Format(entryPkg.FilenameLayout),
						},
					}

					createSubDirectoriedFiles(d, subdirectoriedEntries)

					entries, err := entriesIn(d)
					c.Assume(err, IsNil)

					c.Expect(entries, ContainsExactly, entryFilenames)
				})
			})

			c.Specify("no entries", func() {
				d, cleanUp, err := tmpDir("no_entries")
				c.Assume(err, IsNil)
				defer cleanUp()

				createSomeFiles(d)

				entries, err := entriesIn(d)
				c.Assume(err, IsNil)

				c.Expect(len(entries), Equals, 0)
			})
		})
	})
}
