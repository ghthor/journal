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

// Execute `git add {filepath}` in workingDirectory
func AddFilepath(workingDirectory string, filepath string) error {
	return Command(workingDirectory, "add", filepath).Run()
}

// Execute `git commit -m {msg}` in workingDirectory
func CommitWithMessage(workingDirectory string, msg string) error {
	return Command(workingDirectory, "commit", "-m", msg).Run()
}
