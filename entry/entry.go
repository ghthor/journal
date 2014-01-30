package entry

import (
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
	"os"
	"time"
)

type directory string

type NewEntry interface {
	Open(Now func() time.Time, ideas []idea.Idea) (OpenEntry, error)
}

type OpenEntry interface {
	OpenedAt() time.Time
	Ideas() []idea.Idea

	Edit(mutateIntoEditor func() error) (OpenEntry, error)

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
	return &openEntry{nil, Now(), ideas}, nil
}

type openEntry struct {
	file     *os.File
	openedAt time.Time

	ideas []idea.Idea
}

func (e *openEntry) OpenedAt() time.Time { return e.openedAt }
func (e *openEntry) Ideas() []idea.Idea  { return e.ideas }

func (e *openEntry) Edit(mutateIntoEditor func() error) (OpenEntry, error) {
	return e, nil
}
func (e *openEntry) Close() (ClosedEntry, []idea.Idea, error) {
	return nil, nil, nil
}

type closedEntry struct {
	file     os.File
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
