package entry

import (
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
	"os"
	"path/filepath"
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
	Ideas() []idea.Idea

	Edit(EditorProcess) (OpenEntry, error)

	Close() (ClosedEntry, []idea.Idea, error)
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
func (e *openEntry) Ideas() []idea.Idea  { return e.ideas }

func (e *openEntry) Edit(proc EditorProcess) (OpenEntry, error) {
	err := proc.Start()
	if err != nil {
		return e, err
	}

	return e, proc.Wait()
}
func (e *openEntry) Close() (ClosedEntry, []idea.Idea, error) {
	return nil, nil, nil
}

type closedEntry struct {
	openedAt time.Time
	closedAt time.Time
}

// Implement Commitable
func (e *closedEntry) FilesToAdd() ([]string, error) {
	return nil, nil
}
func (e *closedEntry) CommitMsg() (string, error) {
	return "", nil
}
