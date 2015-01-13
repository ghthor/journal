package main

import (
	"flag"
	"fmt"
	"os"

	verb "github.com/ghthor/journal/cmd_verbs/fix"
)

const (
	EC_OK int = iota
	EC_WD_ERROR
	EC_CMD_ERROR
)

var usagePrefix = `journal-fix updates a journal's file and directory storage format

Usage:
    journal-fix [directory]
`

func main() {
	flagSet := flag.NewFlagSet("journal-fix", flag.ExitOnError)
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
