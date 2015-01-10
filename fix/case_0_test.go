package fix

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghthor/journal/fix/case_0_static"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"

	"code.google.com/p/go.tools/godoc/vfs/mapfs"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

func DescribeFixingCase0(c gospec.Context) {
	tmpDir := func(prefix string) (directory string, cleanUp func()) {
		directory, err := ioutil.TempDir("", prefix+"_")
		c.Assume(err, IsNil)

		cleanUp = func() {
			c.Assume(os.RemoveAll(directory), IsNil)
		}

		return
	}

	c.Specify("case 0 can be fixed", func() {
		baseDir, cleanUp := tmpDir("case_0_can_be_fixed")
		defer cleanUp()

		d, expectedEntries, err := case_0_static.NewIn(baseDir)
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
				c.Expect(string(entry.bytes()), Equals, expectEntries[i])
			}
		})

		c.Specify("and all changes will be commited", func() {
			c.Expect(git.IsClean(d), IsNil)

			vfs := mapfs.New(case_0_static.Files)

			for i, ref := range refLog {
				actual, err := git.Command(d, "show", "--pretty=format:%s%n", ref).Output()
				c.Assume(err, IsNil)

				expectedOutputFilename := filepath.Join("case_0_fix_reflog", fmt.Sprint(i))
				expectedFile, err := vfs.Open(expectedOutputFilename)
				defer expectedFile.Close()
				c.Assume(err, IsNil)

				expected, err := ioutil.ReadAll(expectedFile)
				c.Assume(err, IsNil)

				c.Expect(string(actual), Equals, string(expected))
			}
		})
	})
}
