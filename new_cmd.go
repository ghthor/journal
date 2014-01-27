package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"
)

var editEntry bool
var ignoreDirty bool

func init() {
	flag.BoolVar(&editEntry, "edit", true, "open the target entry in the editor")
	flag.BoolVar(&ignoreDirty, "ignoredirty", false, "ignore if the git repository is dirty")
}

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

		_, err = newEntry(wd, entryTmpl, nil, MutateInto, c, args...)
		if err != nil {
			log.Fatal(err)
		}
	},
}

//A layout to use as the entry's filename
const filenameLayout = "2006-01-02-1504-MST"

func newEntry(dir string, entryTmpl *template.Template, Now func() time.Time, mutateIntoEditor func(*exec.Cmd) (Process, error), c *Command, args ...string) (j journalEntry, err error) {
	if !ignoreDirty {
		if err := GitIsClean(dir); err != nil {
			return j, err
		}
	}

	b := bytes.NewBuffer(make([]byte, 0, 256))

	if Now == nil {
		Now = time.Now
	}
	now := Now()

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
	if mutateIntoEditor != nil && editEntry {
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

	j.ClosedAt = Now().Format(time.UnixDate)

	// Append the ClosedAt time to the file
	f, err := os.OpenFile(entryFilepath, os.O_RDWR, 0600)
	if err != nil {
		return j, err
	}
	defer f.Close()

	// Goto EOF
	_, err = f.Seek(-1, 2)
	if err != nil {
		return j, err
	}

	lastTwoBytes := make([]byte, 2)

	_, err = f.Read(lastTwoBytes)
	if err != nil {
		return j, err
	}

	fbuf := bufio.NewWriter(f)

	switch {
	case bytes.Equal(lastTwoBytes, []byte("\n\n")):
		fbuf.WriteString(j.ClosedAt + "\n")
	default:
		fbuf.WriteString("\n" + j.ClosedAt + "\n")
	}

	err = fbuf.Flush()
	if err != nil {
		return j, err
	}

	// Parse the commit msg from the journal entry
	data, err := ioutil.ReadFile(entryFilepath)
	if err != nil {
		return j, err
	}

	var commitMsg string

	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		line := s.Text()
		if i := strings.Index(line, "#~"); i == 0 {
			commitMsg = line[3:]
		} else {
			continue
		}
	}

	if commitMsg == "" {
		return j, errors.New("entry is missing an event to use as the commit message")
	}

	// Commit the new journal entry to the git repository
	if err := GitAdd(dir, entryFilepath); err != nil {
		return j, err
	}

	if err := GitCommitAll(dir, commitMsg); err != nil {
		return j, err
	}

	return j, nil
}

type journalEntry struct {
	Filename string
	OpenedAt string
	ClosedAt string
}

var entryTmpl = template.Must(template.New("entry").Parse(
	`{{.OpenedAt}}

#~ Title(will be used as commit message)
TODO Make this some random quote or something stupid

## [active] An Idea
An idea carries over from entry to entry if it is active.
`))
