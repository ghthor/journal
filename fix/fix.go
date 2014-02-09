package fix

import (
	"errors"
	"fmt"
	"github.com/ghthor/journal/entry"
	"github.com/ghthor/journal/idea"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type EntriesByDate []string

func (f EntriesByDate) Len() int { return len(f) }
func (f EntriesByDate) Less(i, j int) bool {
	iTime, err := time.Parse(entry.FilenameLayout, f[i])
	if err != nil {
		panic(err)
	}

	jTime, err := time.Parse(entry.FilenameLayout, f[j])
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

	// Move entries to entry/ directory
	entries, err = mvEntriesIn(directory, entries)
	if err != nil {
		return err
	}

	// Initialize an idea directory store
	err = os.Mkdir(filepath.Join(directory, "idea"), 0700)
	if err != nil {
		return err
	}

	ideaStore, _, err := idea.InitDirectoryStore(filepath.Join(directory, "idea"))
	if err != nil {
		return err
	}

	// Store all existing ideas in the directory store
	for i := 0; i < len(entries); i++ {
		entryFile, err := os.OpenFile(filepath.Join(directory, entries[i]), os.O_RDONLY, 0600)
		if err != nil {
			return err
		}
		defer entryFile.Close()

		scanner := idea.NewIdeaScanner(entryFile)
		for scanner.Scan() {
			newIdea := scanner.Idea()

			// Look for an Idea with the same Name
			err = filepath.Walk(filepath.Join(directory, "idea"), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					if info.Name() != "nextid" && info.Name() != "active" {
						ideaFile, err := os.OpenFile(path, os.O_RDONLY, 0660)
						if err != nil {
							return err
						}
						defer ideaFile.Close()

						scanner := idea.NewIdeaScanner(ideaFile)
						if !scanner.Scan() {
							return errors.New(fmt.Sprintf("unexpected file in idea store %s", path))
						}

						idea := scanner.Idea()

						if newIdea.Name == idea.Name {
							newIdea.Id = idea.Id
						}
					}
				}

				return nil
			})

			_, err = ideaStore.SaveIdea(newIdea)
			if err != nil {
				return err
			}
		}
	}

	// Fix entries
	for _, entryFilename := range entries {
		entryFile, err := os.OpenFile(filepath.Join(directory, entryFilename), os.O_RDWR, 0600)
		if err != nil {
			return err
		}
		defer entryFile.Close()

		entry, err := NewEntry(entryFile)
		if err != nil {
			return err
		}

		if entry.NeedsFixed() {
			_, err = entryFile.Seek(0, 0)
			if err != nil {
				return err
			}

			entry, _, err = entry.FixedEntry()
			if err != nil {
				return err
			}

			n, err := io.Copy(entryFile, entry.NewReader())
			if err != nil {
				return err
			}

			err = entryFile.Truncate(n)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
