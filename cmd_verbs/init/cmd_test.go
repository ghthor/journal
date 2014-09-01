package init_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"

	"github.com/ghthor/journal/cmd"
	initVerb "github.com/ghthor/journal/cmd_verbs/init"
	initialize "github.com/ghthor/journal/init"
)

func DescribeInitCmd(c gospec.Context) {
	var tmpDir = func() (directory string, cleanUp func()) {
		directory, err := ioutil.TempDir("", "journal-cmd_verbs-init")
		c.Assume(err, IsNil)

		cleanUp = func() {
			c.Assume(os.RemoveAll(directory), IsNil)
		}

		return
	}

	exec := func(cmd cmd.Cmd, directory string, args []string) {
		c.Assume(initialize.HasBeenInitialized(directory), IsFalse)

		c.Specify("and commit the modifications to git", func() {
			c.Expect(cmd.Exec(args), IsNil)
			c.Expect(initialize.HasBeenInitialized(directory), IsTrue)
			// TODO check commit messages
		})

		c.Specify("and will not commit the modifications to git", func() {
			gitargs := make([]string, 0, len(args)+1)
			gitargs = append(gitargs, "-no-commit")
			gitargs = append(gitargs, args...)

			c.Expect(cmd.Exec(gitargs), IsNil)
			c.Expect(initialize.HasBeenInitialized(directory), IsTrue)
			// TODO check for no new commit messages
		})
	}

	c.Specify("the `init` command", func() {
		d, cleanUp := tmpDir()
		defer cleanUp()

		c.Specify("will initialize a journal", func() {
			c.Specify("using the working directory", func() {
				cmd := initVerb.NewCmd(nil)
				cmd.SetWd(d)
				exec(cmd, d, []string{})
			})

			c.Specify("using a relative path passed as an argument", func() {
				cmd := initVerb.NewCmd(nil)
				cmd.SetWd(d)

				path := filepath.Join(d, "rel/path")
				exec(cmd, path, []string{"rel/path"})
			})

			c.Specify("using the absolute path passed as an argument", func() {
				args := []string{d}
				exec(initVerb.NewCmd(nil), d, args)
			})

		})

		c.Specify("will error with too many arguments", func() {
			cmd := initVerb.NewCmd(nil)

			err := cmd.Exec([]string{d, "another/argument"})
			c.Expect(err, Not(IsNil))
			c.Expect(err.Error(), Equals, "too many arguments")
		})
	})
}
