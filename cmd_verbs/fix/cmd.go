package fix

import (
	"errors"
	"flag"
	"log"
	"path/filepath"

	"github.com/ghthor/journal/fix"
)

var Cmd = NewCmd(nil)

type cmd struct {
	flagSet *flag.FlagSet

	wd string // working directory

	noCommit bool
}

func NewCmd(flagSet *flag.FlagSet) *cmd {
	if flagSet == nil {
		flagSet = flag.NewFlagSet("fix", flag.ExitOnError)
	}

	c := &cmd{
		flagSet: flagSet,
	}

	//c.flagSet.BoolVar(&c.noCommit, "no-commit", false, "don't commit the modifications made by fix to the repository")

	return c
}

func (c *cmd) SetWd(directory string) {
	c.wd = directory
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

	// FIX
	_, err := fix.Fix(path)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (c cmd) Summary() string {
	return "    fix\t\tupgrade the storage format"
}
