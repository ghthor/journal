package main

import (
	"errors"
	"log"
	"os/exec"
)

var gitPath string

func init() {
	var err error
	gitPath, err = exec.LookPath("git")
	if err != nil {
		log.Fatal("git must be installed")
	}
}

func GitCommand(workingDirectory string, args ...string) *exec.Cmd {
	c := exec.Command(gitPath, args...)
	c.Dir = workingDirectory
	return c
}

func GitIsClean(dir string) error {
	c := GitCommand(dir, "status", "-s")

	o, err := c.Output()
	if err != nil {
		return err
	}

	if len(o) != 0 {
		return errors.New("directory is dirty")
	}

	return nil
}

func GitAdd(dir string, filepath string) error {
	return GitCommand(dir, "add", filepath).Run()
}

func GitCommitAll(dir string, msg string) error {
	return GitCommand(dir, "commit", "-m", msg).Run()
}
