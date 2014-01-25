package main

import (
	"os"
	"os/exec"
)

type Process interface {
	Wait() error
}

// Binds std[in|out|err] and calls cmd.Start() then returns
// The caller should call cmd.Wait() to be notified of mutation's exit
func MutateInto(cmd *exec.Cmd) (Process, error) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, cmd.Start()
}
