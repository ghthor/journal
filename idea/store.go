package idea

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghthor/journal/git"
)

// Used to manage idea storage in a directory
type DirectoryStore struct {
	root string
}

// Returned if a directory structure doesn't match
// the required format of an idea storage directory
type InvalidDirectoryStoreError struct {
	Err error
}

func (e InvalidDirectoryStoreError) Error() string {
	return fmt.Sprintf("invalid directory store: %v", e.Err)
}

func IsInvalidDirectoryStoreError(err error) bool {
	_, ok := err.(InvalidDirectoryStoreError)
	return ok
}

func isAnDirectoryStore(d string) error {
	nextIdPath := filepath.Join(d, "nextid")

	data, err := ioutil.ReadFile(nextIdPath)
	if err != nil {
		return InvalidDirectoryStoreError{err}
	}

	var nextAvailableId uint
	n, err := fmt.Fscanf(bytes.NewReader(data), "%d\n", &nextAvailableId)
	if err != nil {
		return InvalidDirectoryStoreError{err}
	}

	if n != 1 {
		return InvalidDirectoryStoreError{errors.New("next available id wasn't found")}
	}

	return nil
}

// Checks that the directory contains the correct files
// to be a DirectoryStore.
// If the directory doesn't contain the require files
// with the expected format this function will return an InvalidDirectoryStoreError.
func NewDirectoryStore(directory string) (*DirectoryStore, error) {
	err := isAnDirectoryStore(directory)
	if err != nil {
		return nil, err
	}

	return &DirectoryStore{directory}, nil
}

// Returned if InitDirectoryStore is called on a directory
// that has already been initialized
var ErrInitOnExistingDirectoryStore = errors.New("init on existing directory store")

// Check that the directory is empty
// and if it is then it initializes an empty
// idea directory store.
func InitDirectoryStore(directory string) (*DirectoryStore, git.Commitable, error) {
	err := isAnDirectoryStore(directory)
	if err == nil {
		return nil, nil, ErrInitOnExistingDirectoryStore
	}

	nextIdCounter := filepath.Join(directory, "nextid")
	err = ioutil.WriteFile(nextIdCounter, []byte("1\n"), 0600)
	if err != nil {
		return nil, nil, err
	}

	activeIndex := filepath.Join(directory, "active")
	err = ioutil.WriteFile(activeIndex, []byte(""), 0600)
	if err != nil {
		return nil, nil, err
	}

	changes := git.NewChangesIn(directory)
	changes.Add(git.ChangedFile("nextid"))
	changes.Add(git.ChangedFile("active"))
	changes.Msg = "idea directory store initialized"

	return &DirectoryStore{directory}, changes, nil
}

// Saves an idea to the directory store and
// returns a commitable containing all changes.
// If the idea does not have an id it will be assigned one.
// If the idea does have an id it will be updated.
func (d DirectoryStore) SaveIdea(idea *Idea) (git.Commitable, error) {
	if idea.Id == 0 {
		return d.saveNewIdea(idea)
	}

	return d.UpdateIdea(*idea)
}

var ErrIdeaExists = errors.New("cannot save a new idea because it already exists")

// Saves an idea that doesn't have an id to the directory and
// returns a commitable containing all changes.
// If the idea is already assigned an id this method will
// return ErrIdeaExists
func (d DirectoryStore) SaveNewIdea(idea *Idea) (git.Commitable, error) {
	if idea.Id != 0 {
		return nil, ErrIdeaExists
	}
	return d.saveNewIdea(idea)
}

// Does not check if the idea has an id
func (d DirectoryStore) saveNewIdea(idea *Idea) (git.Commitable, error) {
	changes := git.NewChangesIn(d.root)

	// Retrieve nextid
	data, err := ioutil.ReadFile(filepath.Join(d.root, "nextid"))
	if err != nil {
		return nil, err
	}

	var nextId uint
	_, err = fmt.Fscan(bytes.NewReader(data), &nextId)
	if err != nil {
		return nil, err
	}

	idea.Id = nextId

	// Increment nextid
	nextId++

	err = ioutil.WriteFile(filepath.Join(d.root, "nextid"), []byte(fmt.Sprintf("%d\n", nextId)), 0600)
	if err != nil {
		return nil, err
	}
	changes.Add(git.ChangedFile("nextid"))

	// write to file
	r, err := NewIdeaReader(*idea)
	if err != nil {
		return nil, err
	}

	ideaFile, err := os.OpenFile(filepath.Join(d.root, fmt.Sprint(idea.Id)), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer ideaFile.Close()

	_, err = io.Copy(ideaFile, r)
	if err != nil {
		return nil, err
	}
	changes.Add(git.ChangedFile(filepath.Base(ideaFile.Name())))

	// If Active, append to active index
	if idea.Status == IS_Active {
		activeIndexFile, err := os.OpenFile(filepath.Join(d.root, "active"), os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		defer activeIndexFile.Close()

		_, err = fmt.Fprintln(activeIndexFile, idea.Id)
		if err != nil {
			return nil, err
		}
		changes.Add(git.ChangedFile("active"))
	}

	changes.Msg = fmt.Sprintf("idea - created - %d", idea.Id)

	return changes, nil
}

var ErrIdeaNotModified = errors.New("the idea was not modified")

// Updates an idea that has already been assigned an id and
// exists in the directory already and
// returns a commitable containing all changes.
// If the idea body wasn't modified this method will
// return ErrIdeaNotModified
func (d DirectoryStore) UpdateIdea(idea Idea) (git.Commitable, error) {
	changes := git.NewChangesIn(d.root)

	data, err := ioutil.ReadFile(filepath.Join(d.root, fmt.Sprint(idea.Id)))
	if err != nil {
		return nil, err
	}

	scanner := NewIdeaScanner(bytes.NewReader(data))

	var ideaOnDisk *Idea
	for scanner.Scan() {
		ideaOnDisk = scanner.Idea()
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	if idea == *ideaOnDisk {
		// No change
		return nil, ErrIdeaNotModified
	}

	// Write to new idea data to file
	ir, err := NewIdeaReader(idea)
	if err != nil {
		return nil, err
	}

	ideaFile, err := os.OpenFile(filepath.Join(d.root, fmt.Sprint(idea.Id)), os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer ideaFile.Close()

	_, err = io.Copy(ideaFile, ir)
	if err != nil {
		return nil, err
	}

	changes.Add(git.ChangedFile(fmt.Sprint(idea.Id)))

	if idea.Status != ideaOnDisk.Status {
		if idea.Status == IS_Active {
			// add the id to the active index
			activeIndex, err := os.OpenFile(filepath.Join(d.root, "active"), os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				return nil, err
			}
			defer activeIndex.Close()

			_, err = fmt.Fprintln(activeIndex, idea.Id)
			if err != nil {
				return nil, err
			}
			changes.Add(git.ChangedFile("active"))

		} else if ideaOnDisk.Status == IS_Active {
			// remove the id from the active index
			activeIndex_RDONLY, err := os.OpenFile(filepath.Join(d.root, "active"), os.O_RDONLY, 0600)
			if err != nil {
				return nil, err
			}
			defer activeIndex_RDONLY.Close()

			// Filled with the unremoved ids
			newIndexBuf := bytes.NewBuffer(make([]byte, 0, 256))

			scanner := bufio.NewScanner(activeIndex_RDONLY)
			for scanner.Scan() {
				var id uint
				_, err := fmt.Fscan(bytes.NewReader(scanner.Bytes()), &id)
				if err != nil {
					return nil, err
				}

				// Filter out the id
				if id != idea.Id {
					_, err := fmt.Fprintln(newIndexBuf, scanner.Text())
					if err != nil {
						return nil, err
					}
				}
			}

			// Write the new index back to the file
			activeIndex_WRONLY, err := os.OpenFile(filepath.Join(d.root, "active"), os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				return nil, err
			}
			defer activeIndex_WRONLY.Close()

			_, err = io.Copy(activeIndex_WRONLY, newIndexBuf)
			if err != nil {
				return nil, err
			}
			changes.Add(git.ChangedFile("active"))
		}
	}

	changes.Msg = fmt.Sprintf("idea - updated - %d", idea.Id)

	return changes, nil
}

func activeIdeasIn(directory string) (activeIds []uint, err error) {
	// Scan in the id's from the index file
	data, err := ioutil.ReadFile(filepath.Join(directory, "active"))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	activeIds = make([]uint, 0, 3)

	for scanner.Scan() {
		var id uint
		_, err := fmt.Fscan(bytes.NewReader(scanner.Bytes()), &id)
		if err != nil {
			return nil, err
		}

		activeIds = append(activeIds, id)
	}
	return
}

// unoptimized
func readIdeaFrom(directory string, id uint) (idea Idea, err error) {
	f, err := os.OpenFile(filepath.Join(directory, fmt.Sprint(id)), os.O_RDONLY, 0600)
	if err != nil {
		return Idea{}, nil
	}
	defer f.Close()

	scanner := NewIdeaScanner(f)
	scanner.Scan()

	if scanner.Err() != nil {
		return Idea{}, scanner.Err()
	}

	idea = *scanner.Idea()
	return idea, nil
}

// Returns a slice of the active ideas from the store
func (d DirectoryStore) ActiveIdeas() (ideas []Idea, err error) {
	activeIds, err := activeIdeasIn(d.root)
	if err != nil {
		return nil, err
	}

	ideas = make([]Idea, 0, len(activeIds))
	for _, id := range activeIds {
		idea, err := readIdeaFrom(d.root, id)
		if err != nil {
			return nil, err
		}

		ideas = append(ideas, idea)
	}

	return ideas, nil
}

// Returns the Idea object stored by the id
func (d DirectoryStore) IdeaById(id uint) (idea Idea, err error) {
	f, err := os.OpenFile(filepath.Join(d.root, fmt.Sprint(id)), os.O_RDONLY, 0600)
	if err != nil {
		return Idea{}, err
	}
	defer f.Close()

	scanner := NewIdeaScanner(f)
	scanner.Scan()

	if scanner.Err() != nil {
		return Idea{}, err
	}

	idea = *scanner.Idea()

	return idea, nil
}
