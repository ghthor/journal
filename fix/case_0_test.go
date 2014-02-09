package fix

import (
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/git/gittest"
	"io/ioutil"
	"os"
	"os/exec"
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

// Copy the journal_cases/case_0/ files to directory
func entriesIn(directory string) (entries []string, err error) {
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if !strings.Contains(filepath.Dir(path), ".git") {
				entries = append(entries, info.Name())
			}
		}
		return nil
	})

	return
}

func newCase0(prefix string) (string, []string, error) {
	// Create a _test/ directory for case_0/
	d, err := ioutil.TempDir("_test", prefix+"_")
	if err != nil {
		return d, nil, err
	}

	// cp -r journal_cases/case_0
	err = exec.Command("cp", "-r", journal_case_0_directory, d).Run()
	if err != nil {
		return d, nil, err
	}

	entries, err := entriesIn(d)
	if err != nil {
		return d, nil, err
	}

	return filepath.Join(d, "case_0"), entries, nil
}

const journal_case_0_directory = "journal_cases/case_0"

// I haven't found a way to store a git repository's
// .git folder in another repository so we have to
// build it during test initialization.
// This function is intended to be called during the TestUnitSpecs()
// function so the cleanupFn can be deferred.
func initCase0() (cleanupFn func(), err error) {
	// Collect the entries we have to commit
	filenames := make([]string, 0, 6)
	err = filepath.Walk(journal_case_0_directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			filenames = append(filenames, info.Name())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// git init
	err = git.Init(journal_case_0_directory)
	if err != nil {
		return nil, err
	}

	// commit the entries
	sort.Sort(entryFilenames(filenames))

	for i, entryFilename := range filenames {
		changes := git.NewChangesIn(journal_case_0_directory)
		changes.Add(git.ChangedFile(entryFilename))
		changes.Msg = fmt.Sprintf("Commit Msg | Entry %d\n", i+1)
		err = changes.Commit()
		if err != nil {
			return nil, err
		}
	}

	// Return a closure that will remove the `journal_cases/case_0/.git` directory
	return func() {
		err := os.RemoveAll(filepath.Join(journal_case_0_directory, ".git"))
		if err != nil {
			panic(err)
		}
	}, nil
}

func DescribeJournalCase0(c gospec.Context) {
	c.Specify("case 0", func() {
		c.Specify("is created as a git repository", func() {
			d, entries, err := newCase0("case_0_init")
			c.Assume(err, IsNil)

			c.Assume(d, gittest.IsAGitRepository)
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
