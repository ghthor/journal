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
	f, err := os.Open(directory)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	contents, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, info := range contents {
		// Ignore Directories
		if info.IsDir() {
			continue
		}

		// Ignore any filesnames that aren't dates in the entry.FilenameLayout
		if _, err := time.Parse(entryPkg.FilenameLayout, info.Name()); err != nil {
			continue
		} else {
			// Collect Entry
			entries = append(entries, info.Name())
		}
	}

	// Recover from panic'ed errors in sort.Sort()
	defer func() {
		if perr := recover(); perr != nil {
			err = perr.(error)
		}
	}()

	// Can Panic, Should never ever panic due to specified behavior
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

	output, err := mvEntries.CombinedOutput()
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("error moving entries to %s : %v", filepath.Join(directory, "entry/"), string(output)))
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

			entry, commitable, err := entry.fixedEntry()
			if err != nil {
				return nil, err
			}

			n, err := io.Copy(entryFile, entry.newReader())
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

// If returns false, then error may or may not be nill
// If returns true, error MUST be nil
func NeedsFixed(directory string) (bool, error) {
	entries, err := entriesIn(directory)
	if err != nil {
		return false, err
	}

	if len(entries) != 0 {
		return true, nil
	}

	if fi, err := os.Stat(filepath.Join(directory, "entry")); os.IsNotExist(err) {
		return true, nil
	} else if !fi.IsDir() {
		return false, errors.New(fmt.Sprintf("%s filesystem node isn't a directory", filepath.Join(directory, "entry")))
	}

	return false, nil
}

func Fix(directory string) (refLog []string, err error) {
	needsFixed, err := NeedsFixed(directory)
	if err != nil {
		return nil, err
	}

	// Drop out early if theres nothing to fix
	if !needsFixed {
		return nil, nil
	}

	// Make a blanket assumption that we're dealing with case0
	return fixCase0(directory)
}
