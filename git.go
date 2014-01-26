package main

import (
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
