package main

import (
	"flag"
	"fmt"
	"github.com/ghthor/journal/config"
	"log"
	"os"
	"strings"
	"text/template"
)

var commands = []*Command{
	newEntryCmd,
}

func main() {
	configPath := flag.String("config", os.ExpandEnv("$HOME/.journal-config.json"), "a path to the configuration file")

	flag.Usage = usage
	flag.Parse()

	if c, err := config.ReadFromFile(*configPath); err == nil {
		if err := os.Chdir(os.ExpandEnv(c.Directory)); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	args := flag.Args()
	if len(args) == 0 || args[0] == "-h" {
		flag.Usage()
		return
	}

	var cmd *Command
	name := args[0]
	for _, c := range commands {
		if strings.HasPrefix(c.Name, name) {
			cmd = c
			break
		}
	}

	if cmd == nil {
		fmt.Printf("error: unknown command %q\n", name)
		flag.Usage()
		os.Exit(1)
	}

	cmd.Exec(args[1:])
}

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
