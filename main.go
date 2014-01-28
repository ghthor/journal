package main

import (
	"flag"
	"fmt"
	"github.com/ghthor/journal/config"
	"log"
	"os"
	"text/template"
)

const (
	EC_OK int = iota
	EC_NO_CMD
	EC_UNKNOWN_COMMAND
)

func usage() {
	fmt.Print(usagePrefix)
	flag.PrintDefaults()
	usageTmpl.Execute(os.Stdout, commands)
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
    {{.Name | printf "%-10s"}} {{.Summary}}{{end}}
`))

func showUsageAndExit(exitCode int) {
	flag.Usage()
	os.Exit(exitCode)
}

var commands = []*Command{
	newEntryCmd,
}

func main() {
	showUsage := flag.Bool("h", false, "show this usage documentation")

	configPath := flag.String("config", os.ExpandEnv("$HOME/.journal-config.json"), "a path to the configuration file")
	init := flag.Bool("init", false, "`git init` the journal directory if it doesn't exist")

	flag.Usage = usage
	flag.Parse()

	if *showUsage {
		showUsageAndExit(EC_OK)
	}

	if c, err := config.ReadFromFile(*configPath); err == nil {
		_, err := os.Stat(os.ExpandEnv(c.Directory))
		if os.IsNotExist(err) && *init {
			err := GitInit(c.Directory)
			if err != nil {
				log.Fatal(err)
			}
		} else if err != nil {
			log.Fatal(err)
		}

		if err := os.Chdir(os.ExpandEnv(c.Directory)); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	var cmd *Command
	args := flag.Args()

	if len(args) == 0 {
		showUsageAndExit(EC_NO_CMD)
	}

	name := args[0]

	for _, c := range commands {
		if c.Name == name {
			cmd = c
			break
		}
	}

	if cmd == nil {
		fmt.Printf("error: unknown command %q\n", name)
		showUsageAndExit(EC_UNKNOWN_COMMAND)
	}

	cmd.Exec(args[1:])
}
