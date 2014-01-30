package entry

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

//A layout to use as the entry's filename
const filenameLayout = "2006-01-02-1504-MST"

var entryTmpl = template.Must(template.New("entry").Parse(
	`{{.OpenedAt}}

#~ Title(will be used as commit message)
TODO Make this some random quote or something stupid
{{range .ActiveIdeas}}
## [{{.Status}}] {{.Name}}
{{.Body}}{{end}}`))

type NewEntry interface {
	Open(Now func() time.Time, ideas []idea.Idea) (OpenEntry, error)
}

type EditorProcess interface {
	Start() error
	Wait() error
}

type OpenEntry interface {
	OpenedAt() time.Time
	Ideas() ([]idea.Idea, error)

	Edit(EditorProcess) (OpenEntry, error)

	Close(time.Time) (ClosedEntry, []idea.Idea, error)
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

func (e *newEntry) Open(Now func() time.Time, ideas []idea.Idea) (OpenEntry, error) {
	openedAt := Now()

	f, err := os.OpenFile(filepath.Join(e.directory, openedAt.Format(filenameLayout)), os.O_WRONLY|os.O_CREATE, 0600)
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
	filename := filepath.Join(e.directory, e.openedAt.Format(filenameLayout))

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

func (e *openEntry) Close(closedAt time.Time) (ClosedEntry, []idea.Idea, error) {
	filename := filepath.Join(e.directory, e.openedAt.Format(filenameLayout))

	f, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	ideas := make([]idea.Idea, 0, len(e.ideas))
	ideaScanner := idea.NewIdeaScanner(f)
	for ideaScanner.Scan() {
		if err := ideaScanner.Err(); err != nil {
			return nil, nil, err
		}
		ideas = append(ideas, *ideaScanner.Idea())
	}

	// Remove the Idea's from the File
	// To Beginning of File
	if _, err = f.Seek(0, 0); err != nil {
		return nil, nil, err
	}

	fr := bufio.NewReader(f)
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	// Find the start of the Idea's list
	// Collect the Lines up till then
	for {
		line, err := fr.ReadString('\n')
		if err != nil {
			return nil, nil, err
		}

		if i := strings.Index(line, "## ["); i != -1 {
			break
		}

		buf.WriteString(line)
	}

	// To Beginning of File
	if _, err = f.Seek(0, 0); err != nil {
		return nil, nil, err
	}

	// Clean up the Trailing newlines
	b := buf.Bytes()
	b = bytes.TrimRight(b, "\n")
	b = append(b, '\n')

	// Write back Contents with Idea's Truncated out
	n, err := f.Write(b)
	if err != nil {
		return nil, nil, err
	}

	if n != len(b) {
		return nil, nil, errors.New("error re-writing entry without idea's")
	}

	closedAtStr := closedAt.Format(time.UnixDate)
	nn, err := f.WriteString("\n" + closedAtStr + "\n")
	if err != nil {
		return nil, nil, err
	}

	// Truncate the File to it's new length
	if err = f.Truncate(int64(n + nn)); err != nil {
		return nil, nil, err
	}

	return &closedEntry{e.directory, e.openedAt, closedAt}, ideas, nil
}

type closedEntry struct {
	directory string

	openedAt time.Time
	closedAt time.Time
}

// Implement Commitable
func (e *closedEntry) FilesToAdd() ([]string, error) {
	return []string{
		filepath.Join(e.directory, e.openedAt.Format(filenameLayout)),
	}, nil
}

func (e *closedEntry) CommitMsg() (string, error) {
	f, err := os.OpenFile(filepath.Join(e.directory, e.openedAt.Format(filenameLayout)),
		os.O_RDONLY, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}

		line := scanner.Text()
		if i := strings.Index(line, "#~"); i != -1 {
			return "ENTRY - " + strings.TrimPrefix(line, "#~ "), nil
		}
	}
	return "", errors.New("missing commit msg")
}
