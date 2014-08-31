package init

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
)

func isAGitRepository(directory string) bool {
	c := git.Command(directory, "status", "-s")

	_, err := c.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

func CanBeInitialized(directory string) (bool, error) {
	di, err := os.Stat(directory)

	// If the directory doesn't exist
	if err != nil && os.IsNotExist(err) {
		return true, nil
	}

	// If the directory exists and is empty
	if !di.IsDir() {
		return false, errors.New(fmt.Sprintf("\"%s\" isn't a directory", directory))
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
		return false, errors.New(fmt.Sprintf("\"%s\" isn't an empty directory", directory))
	}

	return true, nil
}

func entryStoreIsWithin(directory string) bool {
	fi, err := os.Stat(filepath.Join(directory, "entry"))
	if err != nil {
		return false
	}

	return fi.IsDir()
}

func ideaStoreIsWithin(directory string) bool {
	// TODO Implement a clear function to test for an idea.DirectoryStore existence
	_, err := idea.NewDirectoryStore(filepath.Join(directory, "idea"))
	if err != nil {
		return false
	}

	return true
}

func HasBeenInitialized(directory string) bool {
	isGit := isAGitRepository(directory)
	containsEntryStore := entryStoreIsWithin(directory)
	containsIdeaStore := ideaStoreIsWithin(directory)

	return isGit && containsEntryStore && containsIdeaStore
}

func Journal(directory string) error {
	// Check if we need to `git init` the directory
	if !isAGitRepository(directory) {
		err := git.Init(directory)

		if err != nil {
			return err
		}
	}

	// Create the Entry Store
	err := os.Mkdir(filepath.Join(directory, "entry"), 0755)
	if err != nil {
		return err
	}

	// Create the Idea Store
	err = os.Mkdir(filepath.Join(directory, "idea"), 0755)
	if err != nil {
		return err
	}

	_, _, err = idea.InitDirectoryStore(filepath.Join(directory, "idea"))
	if err != nil {
		return err
	}

	return nil
}
