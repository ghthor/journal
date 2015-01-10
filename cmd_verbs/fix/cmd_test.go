package fix_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"

	"github.com/ghthor/journal/cmd"
	"github.com/ghthor/journal/git"

	"github.com/ghthor/journal/cmd_verbs/fix"
	fixPkg "github.com/ghthor/journal/fix"
	"github.com/ghthor/journal/fix/case_0_static"
)

func DescribeFixCmd(c gospec.Context) {
	var tmpDir = func() (directory string, cleanUp func()) {
		directory, err := ioutil.TempDir("", "journal-cmd_verbs-fix")
		c.Assume(err, IsNil)

		cleanUp = func() {
			c.Assume(os.RemoveAll(directory), IsNil)
		}

		return
	}

	exec := func(cmd cmd.Cmd, journalDir string, args []string) {
		needsFixed, err := fixPkg.NeedsFixed(journalDir)
		c.Assume(err, IsNil)
		c.Assume(needsFixed, IsTrue)

		c.Specify("and commit the modifications to git", func() {
			c.Expect(cmd.Exec(args), IsNil)

			needsFixed, err := fixPkg.NeedsFixed(journalDir)
			c.Assume(err, IsNil)
			c.Expect(needsFixed, IsFalse)

			c.Expect(git.IsClean(journalDir), IsNil)
			// TODO compare commits
		})

		// TODO Implement -no-commit flag
		// c.Specify("and will not commit the modifications to git", func() {
		// 	gitargs := make([]string, 0, len(args)+1)
		// 	gitargs = append(gitargs, "-no-commit")
		// 	gitargs = append(gitargs, args...)

		// 	c.Expect(cmd.Exec(gitargs), IsNil)

		// 	needsFixed, err := fixPkg.NeedsFixed(directory)
		// 	c.Assume(err, IsNil)
		// 	c.Expect(needsFixed, IsFalse)

		// 	c.Expect(git.IsClean(directory), Not(IsNil))
		// })
	}

	c.Specify("the `fix` command", func() {
		d, cleanUp := tmpDir()
		defer cleanUp()

		c.Specify("will fix a journal", func() {
			journalDir, _, err := case_0_static.NewIn(d)
			c.Assume(err, IsNil)

			c.Specify("using the working directory", func() {
				cmd := fix.NewCmd(nil)
				cmd.SetWd(journalDir)
				exec(cmd, journalDir, []string{})
			})

			c.Specify("using a relative path passed as an argument", func() {
				cmd := fix.NewCmd(nil)
				cmd.SetWd(d)

				path := filepath.Join(d, "case_0")
				exec(cmd, path, []string{"case_0"})
			})

			c.Specify("using the absolute path passed as an argument", func() {
				args := []string{journalDir}
				exec(fix.NewCmd(nil), journalDir, args)
			})

		})

		c.Specify("will error with too many arguments", func() {
			cmd := fix.NewCmd(nil)

			err := cmd.Exec([]string{d, "another/argument"})
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "too many arguments")
		})
	})
}
