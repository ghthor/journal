package cmd_verbs

import (
	"github.com/ghthor/journal/cmd"
)

// This catalog is private so we can trust that no
// verbs have been overidden by some external package.
var c cmd.Catalog = cmd.NewCatalog()

// Search the cmd.Catalog with a verb
func MatchVerb(verb string) cmd.Cmd {
	return c.MatchVerb(verb)
}

type VerbUsage struct {
	Verb, Summary string
}

// Return a slice of all registered verbs and their usage summaries
func Usages() (usages []VerbUsage) {
	for verb, cmd := range c {
		usages = append(usages, VerbUsage{verb, cmd.Summary()})
	}
	return
}
