package main

import (
	"io"
)

type EntryCase interface {
	Fix() EntryCase
	NewReader() io.Reader
}

func NewEntryCase(r io.Reader) (EntryCase, error) {
	return nil, nil
}

func NewEntryCaseFromFile(filepath string) (EntryCase, error) {
	return nil, nil
}
