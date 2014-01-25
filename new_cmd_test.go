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

func DescribeNewCmd(c gospec.Context) {
	jd, err := tmpGitRepository("journal_test")
	c.Assume(err, IsNil)

	defer func() {
		err := os.RemoveAll(jd)
		c.Assume(err, IsNil)
	}()

	c.Specify("a temporary git repository", func() {
		c.Expect(jd, IsAGitRepository)
	})

	c.Specify("the `new` command", func() {
		c.Specify("will fail", func() {
			c.Specify("if the journal directory has a dirty git repository", func() {
				c.Assume(ioutil.WriteFile(path.Join(jd, "dirty"), []byte("some data"), os.FileMode(0600)), IsNil)
				err := newEntry(jd, false, &Command{})
				c.Expect(err, Not(IsNil))
			})
		})
	})
}
