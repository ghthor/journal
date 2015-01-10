package new

import (
	"flag"
	"fmt"
)

var Cmd = NewCmd(nil)

type cmd struct {
	flagSet *flag.FlagSet

	wd string // working directory

	// noCommit bool
}

func NewCmd(flagSet *flag.FlagSet) *cmd {
	if flagSet == nil {
		flagSet = flag.NewFlagSet("new", flag.ExitOnError)
	}

	c := &cmd{
		flagSet: flagSet,
	}

	//c.flagSet.BoolVar(&c.noCommit, "no-commit", false, "don't commit the new entry to the git repository")

	return c
}

func (c *cmd) SetWd(directory string) {
	c.wd = directory
}

func (c cmd) Exec([]string) error {
	fmt.Println("Executing the command bound to `new` verb")
	return nil
}

func (c cmd) Summary() string {
	return "    new\t\tcreate, edit, and save an entry to a journal"
}
