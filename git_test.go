package main

import (
	"errors"
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func tmpGitRepository(prefix string) (dir string, err error) {
	dir, err = ioutil.TempDir("", prefix)
	if err != nil {
		return "", err
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", err
	}

	gitInitCmd := exec.Command(gitPath, "init", dir)
	err = gitInitCmd.Run()
	if err != nil {
		return "", err
	}

	return
}

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

	pos = gospec.Messagef(fmt.Sprintf("%s directory doesn't exist", path.Join(d, ".git/")), "%s is a git repository", d)
	neg = gospec.Messagef(fmt.Sprintf("%s directory does exist", path.Join(d, ".git/")), "%s is NOT a git repository", d)

	// Check if a .git directory exists
	if info, err := os.Stat(path.Join(d, ".git/")); !os.IsNotExist(err) {
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

func DescribeGitIntegration(c gospec.Context) {
	c.Specify("a temporary git repository can be created", func() {
		d, err := tmpGitRepository("git_integration_test")
		c.Expect(err, IsNil)
		c.Expect(d, IsAGitRepository)
		c.Expect(os.RemoveAll(d), IsNil)
	})
}
