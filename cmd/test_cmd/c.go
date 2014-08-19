package test_cmd

type Cmd struct {
	wasExecuted bool
}

func (c *Cmd) Exec([]string) error {
	c.wasExecuted = true
	return nil
}

func (c *Cmd) Summary() string {
	return "an example implementation of the cmd.Cmd interface"
}

var C *Cmd = &Cmd{}
