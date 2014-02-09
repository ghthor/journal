package fix

import (
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/git/gittest"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//A layout to use as the entry's filename
const filenameLayout = "2006-01-02-1504-MST"

type entryFilenames []string

func (f entryFilenames) Len() int { return len(f) }
func (f entryFilenames) Less(i, j int) bool {
	iTime, err := time.Parse(filenameLayout, f[i])
	if err != nil {
		panic(err)
	}

	jTime, err := time.Parse(filenameLayout, f[j])
	if err != nil {
		panic(err)
	}

	return jTime.After(iTime)
}
func (f entryFilenames) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func DescribeJournalCase0(c gospec.Context) {
	// Copy the journal_cases/case_0/ files to directory
	copyCase0Files := func(to string) (filenames []string) {
		err := filepath.Walk("journal_cases/case_0", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				fromFile, err := os.OpenFile(path, os.O_RDONLY, 0600)
				if err != nil {
					return err
				}
				defer fromFile.Close()

				toFile, err := os.OpenFile(filepath.Join(to, info.Name()), os.O_CREATE|os.O_WRONLY, info.Mode().Perm())
				if err != nil {
					return err
				}
				defer toFile.Close()

				_, err = io.Copy(toFile, fromFile)
				if err != nil {
					return err
				}

				filenames = append(filenames, info.Name())
			}
			return nil
		})

		c.Assume(err, IsNil)

		return
	}

	newCase0 := func(prefix string) (string, []string) {
		// Create a _test/ directory for case_0/
		d, err := ioutil.TempDir("_test", prefix+"_")
		c.Assume(err, IsNil)

		// git init
		c.Assume(git.Init(d), IsNil)

		// Copy case_0/ files
		entries := copyCase0Files(d)
		sort.Sort(entryFilenames(entries))

		// Commit all the entries
		for i, entryFilename := range entries {
			changes := git.NewChangesIn(d)
			changes.Add(git.ChangedFile(entryFilename))
			changes.Msg = fmt.Sprintf("Commit Msg | Entry %d\n", i+1)
			c.Assume(changes.Commit(), IsNil)
		}

		return d, entries
	}

	c.Specify("case 0", func() {
		c.Specify("is created as a git repository", func() {
			d, entries := newCase0("case_0_is_git")

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
				c.Expect(string(o), Equals, "eda50d431c6ffed54ad220b15e5451d4c19d2d02")
			})
		})
	})
}
