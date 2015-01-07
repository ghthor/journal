package fix

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	entryPkg "github.com/ghthor/journal/entry"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
)

type entriesByDate []string

func (f entriesByDate) Len() int { return len(f) }
func (f entriesByDate) Less(i, j int) bool {
	iTime, err := time.Parse(entryPkg.FilenameLayout, f[i])
	if err != nil {
		panic(err)
	}

	jTime, err := time.Parse(entryPkg.FilenameLayout, f[j])
	if err != nil {
		panic(err)
	}

	return jTime.After(iTime)
}
func (f entriesByDate) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

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

	sort.Sort(entriesByDate(entries))

	return
}

func mvEntriesIn(directory string, entries []string) (movedEntries []string, commit git.Commitable, err error) {
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
		return nil, nil, errors.New(fmt.Sprintf("error moving entries to %s : %v", filepath.Join(directory, "entry/"), err))
	}

	changes := git.NewChangesIn(directory)

	// Update filepaths
	movedEntries = entries
	for i, src := range entries {

		dst := filepath.Join("entry", src)

		// will create a rename change in git
		changes.Add(git.ChangedFile(src))
		changes.Add(git.ChangedFile(dst))

		movedEntries[i] = dst
	}

	changes.Msg = "moved all entries to entry/"

	return movedEntries, changes, nil
}

func lastCommitHashIn(directory string) (string, error) {
	o, err := git.Command(directory, "rev-parse", "HEAD").Output()
	return string(bytes.TrimSpace(o)), err
}

type journalFixCommit struct {
	git.Commitable
}

func (c journalFixCommit) CommitMsg() string {
	return "journal - fix - " + c.Commitable.CommitMsg()
}

type journalFixCommitWithSuffix struct {
	git.Commitable
	suffix string
}

func (c journalFixCommitWithSuffix) CommitMsg() string {
	return "journal - fix - " + c.Commitable.CommitMsg() + " - " + c.suffix
}

func fixCase0(directory string) (refLog []string, err error) {
	// Mark the begining of the fix commit log
	err = git.CommitEmpty(directory, "journal - fix - begin")
	if err != nil {
		return nil, err
	}

	beginHash, err := lastCommitHashIn(directory)
	if err != nil {
		return nil, err
	}
	refLog = append(make([]string, 0, 2), beginHash)

	entries, err := entriesIn(directory)
	if err != nil {
		return nil, err
	}

	// Move entries to entry/ directory
	entries, changes, err := mvEntriesIn(directory, entries)
	if err != nil {
		return nil, err
	}

	err = git.Commit(journalFixCommit{changes})
	if err != nil {
		return nil, err
	}

	commitHash, err := lastCommitHashIn(directory)
	if err != nil {
		return nil, err
	}
	refLog = append(refLog, commitHash)

	// Initialize an idea directory store
	err = os.Mkdir(filepath.Join(directory, "idea"), 0700)
	if err != nil {
		return nil, err
	}

	ideaStore, changes, err := idea.InitDirectoryStore(filepath.Join(directory, "idea"))
	if err != nil {
		return nil, err
	}

	err = git.Commit(journalFixCommit{changes})
	if err != nil {
		return nil, err
	}

	commitHash, err = lastCommitHashIn(directory)
	if err != nil {
		return nil, err
	}
	refLog = append(refLog, commitHash)

	// Store all existing ideas in the directory store
	for i := 0; i < len(entries); i++ {
		entryFile, err := os.OpenFile(filepath.Join(directory, entries[i]), os.O_RDONLY, 0600)
		if err != nil {
			return nil, err
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

			changes, err = ideaStore.SaveIdea(newIdea)
			if err != nil {
				return nil, err
			}

			err = git.Commit(journalFixCommitWithSuffix{
				changes,
				"src:" + entries[i],
			})
			if err != nil {
				return nil, err
			}

			commitHash, err = lastCommitHashIn(directory)
			if err != nil {
				return nil, err
			}
			refLog = append(refLog, commitHash)
		}
	}

	// Fix entries
	for _, entryFilename := range entries {
		entryFile, err := os.OpenFile(filepath.Join(directory, entryFilename), os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
		defer entryFile.Close()

		entry, err := newEntry(entryFile)
		if err != nil {
			return nil, err
		}

		if entry.needsFixed() {
			_, err = entryFile.Seek(0, 0)
			if err != nil {
				return nil, err
			}

			entry, commitable, err := entry.FixedEntry()
			if err != nil {
				return nil, err
			}

			n, err := io.Copy(entryFile, entry.NewReader())
			if err != nil {
				return nil, err
			}

			err = entryFile.Truncate(n)
			if err != nil {
				return nil, err
			}

			changes := commitable.(git.Changes)
			changes.Dir = directory
			changes.Add(git.ChangedFile(entryFilename))

			err = git.Commit(journalFixCommitWithSuffix{
				changes,
				entryFilename,
			})
			if err != nil {
				return nil, err
			}

			commitHash, err = lastCommitHashIn(directory)
			if err != nil {
				return nil, err
			}
			refLog = append(refLog, commitHash)
		}
	}

	// Mark the fix completed in the commit log
	err = git.CommitEmpty(directory, "journal - fix - completed")
	if err != nil {
		return nil, err
	}

	completedHash, err := lastCommitHashIn(directory)
	if err != nil {
		return nil, err
	}
	refLog = append(refLog, completedHash)

	return
}
