package init

import (
	"errors"
	"flag"
	"path/filepath"

	"github.com/ghthor/journal/git"
	initialize "github.com/ghthor/journal/init"
)

var Cmd = NewCmd(nil)

type cmd struct {
	flagSet *flag.FlagSet

	wd string // working directory

	noCommit bool
}

func NewCmd(flagSet *flag.FlagSet) *cmd {
	if flagSet == nil {
		flagSet = flag.NewFlagSet("init", flag.ExitOnError)
	}

	c := &cmd{
		flagSet: flagSet,
	}

	c.flagSet.BoolVar(&c.noCommit, "no-commit", false, "don't commit the modifications made by initialization to the repository")

	return c
}

func (c *cmd) SetWd(directory string) {
	c.wd = directory
}

type journalInitCommit struct {
	git.Commitable
}

func (c journalInitCommit) CommitMsg() string {
	return "journal - init - " + c.Commitable.CommitMsg()
}

func (c *cmd) Exec(args []string) error {
	c.flagSet.Parse(args)

	a := c.flagSet.Args()

	var path string

	switch len(a) {
	case 0:
		path = c.wd
	case 1:
		path = a[0]
		if !filepath.IsAbs(path) {
			path = filepath.Join(c.wd, path)
		}

	default:
		return errors.New("too many arguments")
	}

	commitable, err := initialize.Journal(path)
	if err != nil {
		return err
	}

	if !c.noCommit {
		err := git.CommitEmpty(path, "journal - init - begin")
		if err != nil {
			return err
		}

		err = git.Commit(journalInitCommit{commitable})
		if err != nil {
			return err
		}

		err = git.CommitEmpty(path, "journal - init - completed")
		if err != nil {
			return err
		}
	}

	return nil
}

func (c cmd) Summary() string {
	return "how to use `init` verb"
}
