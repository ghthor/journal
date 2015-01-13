package entry

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
)

//A layout to use as the entry's filename
const FilenameLayout = "2006-01-02-1504-MST"

var entryTmpl = template.Must(template.New("entry").Parse(
	`{{.OpenedAt}}

# Title(will be used as commit message)
TODO Make this some random quote or something stupid
{{range .ActiveIdeas}}
## [{{.Status}}] {{.Name}}
{{.Body}}{{end}}`))

type NewEntry interface {
	Open(now time.Time, ideas []idea.Idea) (OpenEntry, error)
}

type EditorProcess interface {
	Start() error
	Wait() error
}

type OpenEntry interface {
	OpenedAt() time.Time
	Ideas() ([]idea.Idea, error)

	Edit(EditorProcess) (OpenEntry, error)

	Close(time.Time) (ClosedEntry, error)
}

type ClosedEntry interface {
	git.Commitable
}

func New(directory string) NewEntry {
	return &newEntry{directory}
}

type newEntry struct {
	directory string
}

func (e *newEntry) Open(openedAt time.Time, ideas []idea.Idea) (OpenEntry, error) {
	f, err := os.OpenFile(filepath.Join(e.directory, openedAt.Format(FilenameLayout)), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	type entry struct {
		OpenedAt    string
		ActiveIdeas []idea.Idea
	}

	err = entryTmpl.Execute(f, entry{
		openedAt.Format(time.UnixDate),
		ideas,
	})
	if err != nil {
		return nil, err
	}

	return &openEntry{e.directory, openedAt, ideas}, nil
}

type openEntry struct {
	directory string

	openedAt time.Time

	ideas []idea.Idea
}

func (e *openEntry) OpenedAt() time.Time { return e.openedAt }
func (e *openEntry) Ideas() ([]idea.Idea, error) {
	filename := filepath.Join(e.directory, e.openedAt.Format(FilenameLayout))

	f, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ideas := make([]idea.Idea, 0, len(e.ideas))
	ideaScanner := idea.NewIdeaScanner(f)
	for ideaScanner.Scan() {
		if err := ideaScanner.Err(); err != nil {
			return nil, err
		}
		ideas = append(ideas, *ideaScanner.Idea())
	}
	e.ideas = ideas
	return ideas, nil
}

func (e *openEntry) Edit(proc EditorProcess) (OpenEntry, error) {
	err := proc.Start()
	if err != nil {
		return e, err
	}

	return e, proc.Wait()
}

var ErrNoCommitMsg = errors.New("entry has no commit msg")

func (e *openEntry) Close(closedAt time.Time) (ClosedEntry, error) {
	filename := filepath.Join(e.directory, e.openedAt.Format(FilenameLayout))

	f, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fr := bufio.NewReader(f)
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	commitMsg := ""

	// Find the start of the Idea's list
	// Collect the Lines up till then
	for {
		line, err := fr.ReadString('\n')
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// Is a commit msg header
		if i := strings.Index(line, "# "); i == 0 {
			commitMsg = strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}

		// Is the Start of an Idea block
		if i := strings.Index(line, "## ["); i != -1 {
			break
		}

		_, err = buf.WriteString(line)
		if err != nil {
			return nil, err
		}
	}

	// Commit Msg Check
	if len(commitMsg) == 0 {
		return nil, ErrNoCommitMsg
	}

	// To Beginning of File
	if _, err = f.Seek(0, 0); err != nil {
		return nil, err
	}

	// Clean up the Trailing newlines
	b := buf.Bytes()
	b = bytes.TrimRight(b, "\n")
	b = append(b, '\n')

	// Write back Contents with Idea's Truncated out
	n, err := f.Write(b)
	if err != nil {
		return nil, err
	}

	if n != len(b) {
		return nil, errors.New("error re-writing entry without idea's")
	}

	closedAtStr := closedAt.Format(time.UnixDate)
	nn, err := f.WriteString("\n" + closedAtStr + "\n")
	if err != nil {
		return nil, err
	}

	// Truncate the File to it's new length
	if err = f.Truncate(int64(n + nn)); err != nil {
		return nil, err
	}

	return &closedEntry{e.directory, commitMsg, e.openedAt, closedAt}, nil
}

type closedEntry struct {
	directory string

	commitMsg string

	openedAt time.Time
	closedAt time.Time
}

func (e *closedEntry) WorkingDirectory() string { return e.directory }
func (e *closedEntry) Changes() []git.CommitableChange {
	return []git.CommitableChange{
		git.ChangedFile(filepath.Join(e.directory, e.openedAt.Format(FilenameLayout))),
	}
}

func (e *closedEntry) CommitMsg() string { return e.commitMsg }
