package new

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

var ErrGitIsDirty = errors.New("git is dirty")

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

func newEnvEditor(envEditor string, entryFilename string) (*exec.Cmd, error) {
	// Enable the $EDITOR variable to
	// contain a string such a "emacs -nw"
	editorArgs := strings.Split(envEditor, " ")

	// Assume that the first item in the split list
	// is the executable name, such as editorArgs[0] == "vim"
	// and look it up.
	editorBin, err := exec.LookPath(editorArgs[0])
	if err != nil {
		return nil, err
	}

	var editorCmd *exec.Cmd

	// Create an *exec.Cmd that will be used to edit the entry
	switch editorArgs[0] {
	case "vim":
		editorCmd = exec.Command(editorBin, "+set spell", entryFilename)
	case "emacs":
		// ignore the "emacs" token from the editorArgs slice
		// and append the entry filename to the end of it
		editorArgs = append(editorArgs[1:], entryFilename)
		editorCmd = exec.Command(editorBin, editorArgs...)
	default:
		// Support for an editor is explicit
		return nil, fmt.Errorf("%v is unimplemented", editorBin)
	}

	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	editorCmd.Stdin = os.Stdin

	return editorCmd, nil
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

	if git.IsClean(path) != nil {
		return ErrGitIsDirty
	}

	// Set default time provider
	if c.Now == nil {
		c.Now = time.Now
	}

	openedAt := c.Now()
	entryFilename := openedAt.Format(entry.FilenameLayout)

	// Define the editor process using the $EDITOR variable
	if c.EditorProcess == nil {
		editorCmd, err := newEnvEditor(os.Getenv("EDITOR"), entryFilename)
		if err != nil {
			return err
		}

		editorCmd.Dir = filepath.Join(path, "entry")

		c.EditorProcess = editorCmd
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
	openEntry, err := entry.Open(openedAt, ideas)
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
		commitable, err := ideaStore.SaveIdea(&i)
		if err != nil {
			if err == idea.ErrIdeaNotModified {
				continue
			}

			return err
		}

		err = git.Commit(commitable)
		if err != nil {
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
	return "create, edit, and save an entry to a journal"
}
