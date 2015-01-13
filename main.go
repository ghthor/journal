package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/ghthor/journal/cmd_verbs"
)

const (
	EC_OK int = iota
	EC_NO_CMD
	EC_CMD_ERROR
	EC_UNKNOWN_COMMAND
	EC_WD_ERROR
	EC_HELP
)

func usage() {
	fmt.Print(usagePrefix)
	usageTmpl.Execute(os.Stdout, cmd_verbs.Usages())
}

var usagePrefix = `journal is a wrapper around git for creating a project/personal log.

Usage:

	journal command [command arguments]

`
var usageTmpl = template.Must(template.New("usage").Parse(
	`The commands are:{{range .}}
    {{.Verb | printf "%-10s"}} {{.Summary}}{{end}}

`))

func showUsageAndExit(exitCode int) {
	flag.Usage()
	os.Exit(exitCode)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	// Check that a verb exists in the arguments
	args := flag.Args()
	if len(args) == 0 {
		showUsageAndExit(EC_NO_CMD)
	}

	// Retrieve the command bound to the verb
	cmd := cmd_verbs.MatchVerb(args[0])
	if cmd == nil {
		if args[0] == "help" {
			showUsageAndExit(EC_HELP)
		}

		fmt.Printf("journal: unknown command `%s`\n", args[0])
		fmt.Println("Run 'journal help' for usage.")
		os.Exit(EC_UNKNOWN_COMMAND)
	}

	// Set Working Directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("error retrieving working directory")
		os.Exit(EC_WD_ERROR)
	}

	cmd.SetWd(wd)

	// Execute the command
	err = cmd.Exec(args[1:])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(EC_CMD_ERROR)
	}
}
