package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
		err := newEntry(true, c, args...)
		if err != nil {
			log.Fatal(err)
		}
	},
}

//A layout to use as the entry's filename
const filenameLayout = "2006-01-02-1504-MST"

func newEntry(openInEditor bool, c *Command, args ...string) error {
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

	err = ioutil.WriteFile(j.Filename, b.Bytes(), os.FileMode(0600))
	if err != nil {
		return err
	}

	// TODO: enable the editor to configurable
	editor, err := exec.LookPath("vim")
	if err != nil {
		return err
	}

	// Open the Editor
	if openInEditor {
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
