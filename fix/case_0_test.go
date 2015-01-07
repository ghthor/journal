package fix

import (
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/git/gittest"
	"github.com/ghthor/journal/idea"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

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
	sort.Sort(entriesByDate(filenames))

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

		c.Specify("can be fixed", func() {
			d, expectedEntries, err := newCase0("case_0_fix")
			c.Assume(err, IsNil)

			refLog, err := fixCase0(d)
			c.Assume(err, IsNil)

			c.Specify("by moving entries into `entry/`", func() {
				info, err := os.Stat(filepath.Join(d, "entry"))
				c.Expect(err, IsNil)
				c.Expect(info.IsDir(), IsTrue)

				actualEntries, err := entriesIn(filepath.Join(d, "entry"))
				c.Assume(err, IsNil)

				c.Expect(actualEntries, ContainsExactly, expectedEntries)
			})

			c.Specify("by storing ideas in `idea/` directory store", func() {
				info, err := os.Stat(filepath.Join(d, "idea"))
				c.Expect(err, IsNil)
				c.Expect(info.IsDir(), IsTrue)

				_, err = idea.NewDirectoryStore(filepath.Join(d, "idea"))
				c.Expect(err, IsNil)

				nextid, err := ioutil.ReadFile(filepath.Join(d, "idea/nextid"))
				c.Assume(err, IsNil)
				c.Expect(string(nextid), Equals, "4\n")

				active, err := ioutil.ReadFile(filepath.Join(d, "idea/active"))
				c.Assume(err, IsNil)
				c.Expect(string(active), Equals, "3\n")
			})

			c.Specify("by fixing formatting differences in all entries", func() {
				expectEntries := []string{
					`Wed Jan  1 00:00:00 EST 2014

# Commit Msg | Entry 1
Entry Body

entry_case_0

Wed Jan  1 00:02:00 EST 2014
`,
					`Thu Jan  2 00:00:00 EST 2014

# Commit Msg | Entry 2
Entry Body

entry_case_1

Thu Jan  2 00:01:00 EST 2014
`,
					`Fri Jan  3 00:00:00 EST 2014

# Commit Msg | Entry 3
Entry Body

entry_case_2

Fri Jan  3 00:01:00 EST 2014
`,
					`Sat Jan  4 00:00:00 EST 2014

# Commit Msg | Entry 4
Entry Body

entry_case_3

Sat Jan  4 00:01:00 EST 2014
`,
					`Sun Jan  5 00:00:00 EST 2014

# Commit Msg | Entry 5
Entry Body

entry_case_4

Sun Jan  5 00:01:00 EST 2014
`,
					`Mon Jan  6 00:00:00 EST 2014

# Commit Msg | Entry 6
Entry Body

entry_case_4

Mon Jan  6 00:01:00 EST 2014
`}

				for i, entryFilename := range expectedEntries {
					entry, err := newEntryFromFile(filepath.Join(d, "entry", entryFilename))
					c.Assume(err, IsNil)
					c.Expect(entry.needsFixed(), IsFalse)
					c.Expect(string(entry.Bytes()), Equals, expectEntries[i])
				}
			})

			c.Specify("and all changes will be commited", func() {
				c.Expect(git.IsClean(d), IsNil)

				for i, ref := range refLog {
					actual, err := git.Command(d, "show", "--pretty=format:%s%n", ref).Output()
					c.Assume(err, IsNil)

					expectedOutputFilename := filepath.Join("journal_cases/case_0_fix_reflog", fmt.Sprint(i))
					expected, err := ioutil.ReadFile(expectedOutputFilename)
					c.Assume(err, IsNil)

					c.Expect(string(actual), Equals, string(expected))
				}
			})
		})
	})
}
