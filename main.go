package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/ghthor/journal/cmd_verbs"
	"github.com/ghthor/journal/config"
)

const (
	EC_OK int = iota
	EC_NO_CMD
	EC_CMD_ERROR
	EC_UNKNOWN_COMMAND
	EC_HELP
)

func usage() {
	fmt.Print(usagePrefix)
	flag.PrintDefaults()
	usageTmpl.Execute(os.Stdout, cmd_verbs.Usages())
}

var usagePrefix = `
journal is a filesystem text based journal that stores metadata about each entry

Usage:
    journal [options] <subcommand> [subcommand options]

Options:
`
var usageTmpl = template.Must(template.New("usage").Parse(
	`
Commands:{{range .}}
    {{.Verb | printf "%-10s"}} {{.Summary}}{{end}}
`))

func showUsageAndExit(exitCode int) {
	flag.Usage()
	os.Exit(exitCode)
}

func main() {
	showUsage := flag.Bool("h", false, "show this usage documentation")

	configPath := flag.String("config", "$HOME/.journal-config.json", "a path to the configuration file")

	flag.Usage = usage
	flag.Parse()

	// Show Help
	if *showUsage {
		showUsageAndExit(EC_HELP)
	}

	*configPath = os.ExpandEnv(*configPath)

	// Open the Config file
	config, err := config.ReadFromFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Check if Directory exists
	_, err = os.Stat(config.Directory)
	if os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Check for Stat error that isn't os.IsNotExist()
	if err != nil {
		log.Fatal(err)
	}

	// Change Working Directory to the one in the configuration file
	if err := os.Chdir(config.Directory); err != nil {
		log.Fatal(err)
	}

	// Check that a verb exists in the arguments
	args := flag.Args()
	if len(args) == 0 {
		showUsageAndExit(EC_NO_CMD)
	}

	// Retrieve the command bound to the verb
	cmd := cmd_verbs.MatchVerb(args[0])
	if cmd == nil {
		showUsageAndExit(EC_UNKNOWN_COMMAND)
	}

	// Execute the command
	err = cmd.Exec(args[1:])
	if err != nil {
		showUsageAndExit(EC_CMD_ERROR)
	}
}
