package git

import (
	"errors"
	"fmt"
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

// Construct an *exec.Cmd for `git {args}` with a workingDirectory
func Command(workingDirectory string, args ...string) *exec.Cmd {
	c := exec.Command(gitPath, args...)
	c.Dir = workingDirectory
	return c
}

// Execute `git init {directory}` in the current workingDirectory
func Init(directory string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return Command(wd, "init", directory).Run()
}

// Execute `git status -s` in directory
// If there is output, the directory has is dirty
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

// Execute `git add --all {filepath}` in workingDirectory
func AddFilepath(workingDirectory string, filepath string) error {
	o, err := Command(workingDirectory, "add", "--all", filepath).CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("error during `git add`: %s\n%s", err.Error(), string(o)))
	}
	return nil
}

// Execute `git commit -m {msg}` in workingDirectory
func CommitWithMessage(workingDirectory string, msg string) error {
	o, err := Command(workingDirectory, "commit", "-m", msg).CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("error during `git commit`: %s\n%s", err.Error(), string(o)))
	}

	return nil
}

// Execute `git commit --allow-empty -m {msg}` in workingDirectory.
func CommitEmpty(workingDirectory string, msg string) error {
	return Command(workingDirectory, "commit", "--allow-empty", "-m", msg).Run()
}
