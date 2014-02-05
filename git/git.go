package git

import (
	"errors"
	"log"
	"os"
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

// Make an *exec.Cmd for `git` with args
func Command(workingDirectory string, args ...string) *exec.Cmd {
	c := exec.Command(gitPath, args...)
	c.Dir = workingDirectory
	return c
}

// `git init` a directory
func Init(directory string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return Command(wd, "init", directory).Run()
}

// Check a directory for staged or un-staged changes
func IsClean(directory string) error {
	c := Command(directory, "status", "-s")

	o, err := c.Output()
	if err != nil {
		return err
	}

	if len(o) != 0 {
		return errors.New("directory is dirty")
	}

	return nil
}

// `git add` a filepath
func AddFilepath(workingDirectory string, filepath string) error {
	return Command(workingDirectory, "add", filepath).Run()
}

func GitCommit(dir string, msg string) error {
	return Command(dir, "commit", "-m", msg).Run()
}
