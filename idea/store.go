package idea

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ghthor/journal/git"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
	changes.Msg = "directory store initialized"

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

	changes.Msg = fmt.Sprintf("IDEA - %d - Created", idea.Id)

	return changes, nil
}

var ErrIdeaNotModified = errors.New("the idea was not modified")

// Updates an idea that has already been assigned an id and
// exists in the directory already and
// returns a commitable containing all changes.
// If the idea body wasn't modified this method will
// return ErrIdeaNotModified
func (d DirectoryStore) UpdateIdea(idea Idea) (git.Commitable, error) {
	return nil, nil
}
