package main

import (
	"flag"
)

type Command struct {
	Run  func(cmd *Command, args ...string) error
	Flag flag.FlagSet

	Name    string
	Summary string
}

func (c *Command) Exec(args []string) error {
	c.Flag.Parse(args)
	return c.Run(c, c.Flag.Args()...)
}
