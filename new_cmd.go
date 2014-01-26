package main

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
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

		_, err = newEntry(wd, entryTmpl, MutateInto, c, args...)
		if err != nil {
			log.Fatal(err)
		}
	},
}

//A layout to use as the entry's filename
const filenameLayout = "2006-01-02-1504-MST"

func IsJournalDirectoryClean(dir string) error {
	c := GitCommand(dir, "status", "-s")

	o, err := c.Output()
	if err != nil {
		return err
	}

	if len(o) != 0 {
		return errors.New("directory is dirty")
	}

	return nil
}

func newEntry(dir string, entryTmpl *template.Template, mutateIntoEditor func(*exec.Cmd) (Process, error), c *Command, args ...string) (j journalEntry, err error) {
	if err := IsJournalDirectoryClean(dir); err != nil {
		return j, err
	}

	b := bytes.NewBuffer(make([]byte, 0, 256))

	now := time.Now()

	j = journalEntry{
		Filename: now.Format(filenameLayout),
		OpenedAt: now.Format(time.UnixDate),
	}

	// *sigh* can't stop laughing.....
	err = entryTmpl.Execute(b, j)
	if err != nil {
		return j, err
	}

	entryFilepath := path.Join(dir, j.Filename)
	err = ioutil.WriteFile(entryFilepath, b.Bytes(), os.FileMode(0600))
	if err != nil {
		return j, err
	}

	// Open the Editor
	if mutateIntoEditor != nil {
		// TODO: enable the editor to configurable
		editorPath, err := exec.LookPath("vim")
		if err != nil {
			return j, err
		}

		editor, err := mutateIntoEditor(exec.Command(editorPath, entryFilepath))
		if err != nil {
			return j, err
		}

		err = editor.Wait()
		if err != nil {
			return j, err
		}
	}

	j.ClosedAt = time.Now().Format(time.UnixDate)

	f, err := os.OpenFile(entryFilepath, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return j, err
	}
	defer f.Close()

	fbuf := bufio.NewWriter(f)
	fbuf.WriteString("\n" + j.ClosedAt + "\n")

	return j, fbuf.Flush()
}

type journalEntry struct {
	Filename string
	OpenedAt string
	ClosedAt string
}

var entryTmpl = template.Must(template.New("entry").Parse(
	`{{.OpenedAt}}

# Subject
TODO Make this some random quote or something stupid
`))
