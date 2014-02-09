package fix

import (
	"errors"
	"fmt"
	"github.com/ghthor/journal/idea"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//A layout to use as the entry's filename
const filenameLayout = "2006-01-02-1504-MST"

type EntriesByDate []string

func (f EntriesByDate) Len() int { return len(f) }
func (f EntriesByDate) Less(i, j int) bool {
	iTime, err := time.Parse(filenameLayout, f[i])
	if err != nil {
		panic(err)
	}

	jTime, err := time.Parse(filenameLayout, f[j])
	if err != nil {
		panic(err)
	}

	return jTime.After(iTime)
}
func (f EntriesByDate) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func entriesIn(directory string) (entries []string, err error) {
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if !strings.Contains(filepath.Dir(path), ".git") {
				entries = append(entries, info.Name())
			}
		}
		return nil
	})

	sort.Sort(EntriesByDate(entries))

	return
}

func mvEntriesIn(directory string, entries []string) (movedEntries []string, err error) {
	err = os.Mkdir(filepath.Join(directory, "entry"), 0700)
	if err != nil {
		return
	}

	mvArgs := make([]string, 0, len(entries)+1)
	mvArgs = append(mvArgs, entries...)
	mvArgs = append(mvArgs, "entry/")

	mvPath, err := exec.LookPath("mv")
	if err != nil {
		return
	}

	mvEntries := exec.Command(mvPath, mvArgs...)
	mvEntries.Dir = directory

	err = mvEntries.Run()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error moving entries to %s : %v", filepath.Join(directory, "entry/"), err))
	}

	// Update filepaths
	movedEntries = entries
	for i, entry := range entries {
		movedEntries[i] = filepath.Join("entry", entry)
	}
	return
}

func FixCase0(directory string) error {
	entries, err := entriesIn(directory)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return errors.New(fmt.Sprintf("%s contains no entries", directory))
	}

	entries, err = mvEntriesIn(directory, entries)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(directory, "idea"), 0700)
	if err != nil {
		return err
	}

	_, _, err = idea.InitDirectoryStore(filepath.Join(directory, "idea"))
	if err != nil {
		return err
	}

	return nil
}
