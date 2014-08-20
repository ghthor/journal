package init

import (
	"os"
	"path/filepath"
	"strings"
)

func CanBeInitialized(directory string) bool {
	di, err := os.Stat(directory)

	// If the directory doesn't exist
	if err != nil && os.IsNotExist(err) {
		return true
	}

	// If the directory exists and is empty
	if !di.IsDir() {
		return false
	}

	numChildren := 0
	filepath.Walk(directory, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory
		if fi.Name() == di.Name() {
			return nil
		}

		// Skip a .git repository
		if strings.Contains(path, ".git") {
			return nil
		}

		numChildren++

		return nil
	})

	if numChildren > 0 {
		return false
	}

	return true
}

func HasBeenInitialized(directory string) bool {
	return false
}

func Journal(directory string) error {
	return nil
}
