package main

import (
	"io"
)

type Entry interface {
	// Needs Fixed
	NeedsFixed() bool

	// Return an Entry that has been fixed
	FixedEntry() Entry

	// Returns a byte slice of the entry w/o fixes
	Bytes() []byte

	// Returns an io.Reader for the entry w/o fixes
	NewReader() io.Reader
}

// A fix for an entry
type EntryFix interface {
	// Returns byte slice that has been fixed
	Fix(io.Reader) []byte
}

func NewEntry(r io.Reader) (Entry, error) {
	return nil, nil
}

func NewEntryFromFile(filepath string) (Entry, error) {
	return nil, nil
}
