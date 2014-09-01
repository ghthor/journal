package main

import (
	"flag"
	"fmt"
	"os"

	verb "github.com/ghthor/journal/cmd_verbs/init"
)

const (
	EC_OK int = iota
	EC_WD_ERROR
	EC_CMD_ERROR
)

var usagePrefix = `
journal-init initializes a directory as a journal data store

Usage:
    journal-init [options] [directory]

Options:
`

func main() {
	flagSet := flag.NewFlagSet("journal-init", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Print(usagePrefix)
		flagSet.PrintDefaults()
	}

	cmd := verb.NewCmd(flagSet)

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(EC_WD_ERROR)
	}

	cmd.SetWd(wd)

	// Execute the command
	err = cmd.Exec(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(EC_CMD_ERROR)
	}
}
