package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func lastEntryFilename(journalDirectory string) (filename string, err error) {
	var last time.Time = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

	err = filepath.Walk(journalDirectory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.Index(path, ".git") == -1 {
			t, err := time.Parse(filenameLayout, filepath.Base(path))
			if err == nil {
				if t.After(last) {
					last = t
					filename = filepath.Base(path)
				}
			}
		}
		return nil
	})

	if filename == "" {
		err = errors.New("journal is empty")
	}

	return
}
