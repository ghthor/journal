package new

import (
	"fmt"
)

type cmd struct {
}

func (c cmd) Exec([]string) error {
	fmt.Println("Executing the command bound to `new` verb")
	return nil
}

func (c cmd) Summary() string {
	return "how to use `new` verb"
}

var Cmd = &cmd{}
