package cmd_verbs

import (
	// init is a reserved keyword
	initc "github.com/ghthor/journal/cmd_verbs/init"

	"github.com/ghthor/journal/cmd_verbs/fix"

	// new is a reserved keyword
	newc "github.com/ghthor/journal/cmd_verbs/new"
)

func init() {
	c.RegisterAsPkg(initc.Cmd)
	c.RegisterAsPkg(fix.Cmd)
	c.RegisterAsPkg(newc.Cmd)
}
