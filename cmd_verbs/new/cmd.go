package new

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ghthor/journal/entry"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
)

var Cmd = NewCmd(nil)

type cmd struct {
	EditorProcess entry.EditorProcess
	Now           func() time.Time

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

func (c cmd) Exec(args []string) error {
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

	// Set defaults
	if c.EditorProcess == nil {
		// Set Up Vim
	}

	if c.Now == nil {
		c.Now = time.Now
	}

	// Make a new entry
	entry := entry.New(filepath.Join(path, "entry"))

	ideaStore, err := idea.NewDirectoryStore(filepath.Join(path, "idea"))
	if err != nil {
		return err
	}

	ideas, err := ideaStore.ActiveIdeas()
	if err != nil {
		return err
	}

	// Open entry w/ ideas
	openEntry, err := entry.Open(c.Now(), ideas)
	if err != nil {
		return err
	}

	// Start editor
	openEntry, err = openEntry.Edit(c.EditorProcess)
	if err != nil {
		return fmt.Errorf("error during edit: %s", err)
	}

	// Parse out the ideas
	ideas, err = openEntry.Ideas()
	if err != nil {
		return err
	}

	// Save the ideas to the store
	for _, i := range ideas {
		_, err := ideaStore.SaveIdea(&i)
		if err != nil {
			if err == idea.ErrIdeaNotModified {
				continue
			}

			return err
		}
	}

	// Save the entry and commit it
	closedEntry, err := openEntry.Close(c.Now())
	if err != nil {
		return err
	}

	err = git.Commit(closedEntry)
	if err != nil {
		return err
	}

	return nil
}

func (c cmd) Summary() string {
	return "    new\t\tcreate, edit, and save an entry to a journal"
}
