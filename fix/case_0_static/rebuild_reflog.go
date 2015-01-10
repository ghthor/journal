// +build ignore

// Command that will rebuild the reflog of the changes that will
// be committed while fixing the journal directory.
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ghthor/journal/fix"
	"github.com/ghthor/journal/fix/case_0_static"
	"github.com/ghthor/journal/git"
)

var wd string

func init() {
	var err error
	wd, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	tmpDir, err := ioutil.TempDir("", "case_0_rebuild_reflog")
	if err != nil {
		log.Fatal(err)
	}

	journalDir, _, err := case_0_static.NewIn(tmpDir)
	if err != nil {
		log.Fatal(err)
	}

	reflog, err := fix.Fix(journalDir)
	if err != nil {
		log.Fatal(err)
	}

	for i, ref := range reflog {
		refFilename := filepath.Join(wd, "case_0_fix_reflog", fmt.Sprint(i))

		err := writeRefToFile(journalDir, ref, refFilename)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func writeRefToFile(journalDir, hash, filename string) error {
	actual, err := git.Command(journalDir, "show", "--pretty=format:%s%n", hash).Output()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewReader(actual))
	if err != nil {
		return err
	}

	return nil
}
