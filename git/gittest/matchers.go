// This package provides GoSpec Matcher's that are useful
// when working with git repositories
package gittest

import (
	"errors"
	"fmt"
	"github.com/ghthor/gospec"
	"os"
	"path/filepath"
)

// TODO: Specify the behavior of this matcher
func IsAGitRepository(dir interface{}, _ interface{}) (match bool, pos gospec.Message, neg gospec.Message, err error) {
	d, ok := dir.(string)
	if !ok {
		return false, pos, neg, errors.New("directory is not a string")
	}

	// Check if jd exists and is a Directory
	if info, err := os.Stat(d); !os.IsNotExist(err) {
		if !info.IsDir() {
			return false, pos, neg, errors.New(fmt.Sprintf("%s is not a directory", d))
		}
	} else {
		// jd doesn't exist
		return false, pos, neg, err
	}

	pos = gospec.Messagef(fmt.Sprintf("%s directory doesn't exist", filepath.Join(d, ".git/")), "%s is a git repository", d)
	neg = gospec.Messagef(fmt.Sprintf("%s directory does exist", filepath.Join(d, ".git/")), "%s is NOT a git repository", d)

	// Check if a .git directory exists
	if info, err := os.Stat(filepath.Join(d, ".git/")); !os.IsNotExist(err) {
		if !info.IsDir() {
			return false, pos, neg, nil
		}
	} else {
		// .git directory doesn't exist
		return false, pos, neg, nil
	}

	pos = gospec.Messagef(d, "%s is a git repository", d)
	neg = gospec.Messagef(d, "%s is NOT a git repository", d)

	match = true
	return
}
