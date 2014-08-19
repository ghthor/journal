package cmd

import (
	"errors"
	"fmt"
	"path"
	"reflect"
)

// An interface to an executable command
type Cmd interface {
	Exec(args []string) error
	Summary() string
}

// A map of verbs -> Command interfaces
type Catalog map[string]Cmd

func NewCatalog() Catalog {
	return make(map[string]Cmd)
}

// Register a verb to a Command
func (c Catalog) Register(verb string, cmd Cmd) error {
	if v, exists := c[verb]; exists {
		return errors.New(fmt.Sprintf("verb already registered: %s -> %s", verb, v))
	}

	c[verb] = cmd
	return nil
}

func (c Catalog) RegisterAsPkg(cmd Cmd) error {
	rv := reflect.ValueOf(cmd)
	return c.Register(path.Base(rv.Elem().Type().PkgPath()), cmd)
}

// Perform a catalog[verb] key value lookup
func (c Catalog) MatchVerb(verb string) Cmd {
	return c[verb]
}

func (c Catalog) MatchExists(verb string) (Cmd, bool) {
	cmd, exists := c[verb]
	return cmd, exists
}
