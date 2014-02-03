package main

import (
	"github.com/ghthor/journal/git"
	"io/ioutil"
)

func tmpGitRepository(prefix string) (dir string, err error) {
	dir, err = ioutil.TempDir("", prefix)
	if err != nil {
		return "", err
	}

	return dir, git.GitInit(dir)
}
