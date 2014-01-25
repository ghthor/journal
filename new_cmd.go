package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
	"text/template"
	"time"
)

var newEntryCmd = &Command{
	Name:    "new",
	Usage:   "",
	Summary: "Create a new journal entry",
	Help:    "TODO",
	Run: func(c *Command, args ...string) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		err = newEntry(wd, true, c, args...)
		if err != nil {
			log.Fatal(err)
		}
	},
}

//A layout to use as the entry's filename
const filenameLayout = "2006-01-02-1504-MST"

func IsJournalDirectoryClean(dir string) error {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return err
	}

	c := exec.Command(gitPath, "status", "-s")
	c.Dir = dir

	o, err := c.Output()
	if err != nil {
		return err
	}

	if len(o) != 0 {
		return errors.New("directory is dirty")
	}

	return nil
}

func newEntry(dir string, mutateIntoEditor bool, c *Command, args ...string) error {
	if err := IsJournalDirectoryClean(dir); err != nil {
		return err
	}

	b := bytes.NewBuffer(make([]byte, 0, 256))

	now := time.Now()

	j := journalEntry{
		Filename:  now.Format(filenameLayout),
		TimeStamp: now.Format(time.UnixDate),
	}

	// *sigh* can't stop laughing.....
	err := entryTmpl.Execute(b, j)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(dir, j.Filename), b.Bytes(), os.FileMode(0600))
	if err != nil {
		return err
	}

	// Open the Editor
	if mutateIntoEditor {
		// TODO: enable the editor to configurable
		editor, err := exec.LookPath("vim")
		if err != nil {
			return err
		}

		// Mutate the Process into the Editor
		err = syscall.Exec(editor, []string{editor, j.Filename}, os.Environ())
		if err != nil {
			return err
		}
	}

	return nil
}

type journalEntry struct {
	Filename  string
	TimeStamp string
}

var entryTmpl = template.Must(template.New("entry").Parse(
	`{{.TimeStamp}}

# Subject
TODO Make this some random quote or something stupid
`))
