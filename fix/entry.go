package fix

import (
	"bytes"
	"github.com/ghthor/journal/git"
	"io"
	"io/ioutil"
	"os"
)

type entry interface {
	// Needs Fixed
	NeedsFixed() bool

	// Return an Entry that has been fixed
	FixedEntry() (entry, git.Commitable, error)

	// Returns a byte slice of the entry w/o fixes
	Bytes() []byte

	// Returns an io.Reader for the entry w/o fixes
	NewReader() io.Reader
}

type entryCaseNeedsFixed struct {
	bytes []byte
	fixes []EntryFix
}

func (e entryCaseNeedsFixed) NeedsFixed() bool { return len(e.fixes) > 0 }
func (e entryCaseNeedsFixed) FixedEntry() (entry, git.Commitable, error) {
	var (
		data []byte = e.bytes
		err  error
	)

	for _, fix := range e.fixes {
		data, err = fix.Execute(bytes.NewReader(data))
		if err != nil {
			return nil, nil, err
		}
	}

	return entryCaseCurrent{data}, git.Changes{
		Msg: "entry - format updated",
	}, nil
}

func (e entryCaseNeedsFixed) Bytes() []byte { return e.bytes }
func (e entryCaseNeedsFixed) NewReader() io.Reader {
	return bytes.NewReader(e.bytes)
}

type entryCaseCurrent struct {
	bytes []byte
}

func (e entryCaseCurrent) NeedsFixed() bool { return false }
func (e entryCaseCurrent) FixedEntry() (entry, git.Commitable, error) {
	return e, nil, nil
}
func (e entryCaseCurrent) Bytes() []byte        { return e.bytes }
func (e entryCaseCurrent) NewReader() io.Reader { return bytes.NewReader(e.bytes) }

func findErrorsInEntry(r io.Reader) (fixes []EntryFix, err error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	for _, fix := range entryFixes {
		if needsFixed, err := fix.CanFix(bytes.NewReader(data)); err != nil {
			return nil, err
		} else if needsFixed {
			fixes = append(fixes, fix)
		}
	}
	return
}

func newEntry(r io.Reader) (entry, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if fixes, err := findErrorsInEntry(bytes.NewReader(data)); err != nil {
		return nil, err
	} else if fixes != nil {
		return entryCaseNeedsFixed{data, fixes}, nil
	}

	return entryCaseCurrent{data}, nil
}

func NewEntryFromFile(filepath string) (entry, error) {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return newEntry(f)
}
