// This package provides GoSpec Matcher's that are useful
// when working with git repositories
package gittest

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ghthor/gospec"
	"os"
	"os/exec"
	"path/filepath"
)

func IsAGitRepository(dir interface{}, _ interface{}) (match bool, pos gospec.Message, neg gospec.Message, err error) {
	d, ok := dir.(string)
	if !ok {
		return false, pos, neg, errors.New("directory is not a string")
	}

	// Check if d exists and is a Directory
	if info, err := os.Stat(d); !os.IsNotExist(err) {
		if !info.IsDir() {
			return false, pos, neg, errors.New(fmt.Sprintf("%s is not a directory", d))
		}
	} else {
		// d doesn't exist
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

func IsInsideAGitRepository(dir interface{}, _ interface{}) (match bool, pos, neg gospec.Message, err error) {
	d, ok := dir.(string)
	if !ok {
		return false, pos, neg, errors.New("directory is not a string")
	}

	// Check if d exists and is a Directory
	if info, err := os.Stat(d); !os.IsNotExist(err) {
		if !info.IsDir() {
			return false, pos, neg, errors.New(fmt.Sprintf("%s is not a directory", d))
		}
	} else {
		// d doesn't exist
		return false, pos, neg, err
	}

	pos = gospec.Messagef(match, "%s is inside a git repository", d)
	neg = gospec.Messagef(match, "%s is not inside a git repository", d)

	git, err := exec.LookPath("git")
	if err != nil {
		return
	}

	cmd := exec.Command(git, "status")
	cmd.Dir = d

	o, _ := cmd.CombinedOutput()

	match = !bytes.Contains(o, []byte("fatal: Not a git repository"))

	return
}
