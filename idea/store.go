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
type IdeaDirectory struct {
	root string
}

// Returned if a directory structure doesn't match
// the required format of an idea storage directory
type InvalidIdeaDirectoryError struct {
	Err error
}

func (e InvalidIdeaDirectoryError) Error() string {
	return fmt.Sprintf("invalid idea directory: %v", e.Err)
}

func IsInvalidIdeaDirectoryError(err error) bool {
	_, ok := err.(InvalidIdeaDirectoryError)
	return ok
}

func isAnIdeaDirectory(d string) error {
	nextIdPath := filepath.Join(d, "nextid")

	data, err := ioutil.ReadFile(nextIdPath)
	if err != nil {
		return InvalidIdeaDirectoryError{err}
	}

	var nextAvailableId uint
	n, err := fmt.Fscanf(bytes.NewReader(data), "%d\n", &nextAvailableId)
	if err != nil {
		return InvalidIdeaDirectoryError{err}
	}

	if n != 1 {
		return InvalidIdeaDirectoryError{errors.New("next available id wasn't found")}
	}

	return nil
}

// Checks that the directory contains the correct files
// to be an IdeaDirectory.
// If the directory doesn't contain the require  files
// with the expected format this function will return an ErrInvalidIdeaDirectory.
func NewIdeaDirectory(directory string) (*IdeaDirectory, error) {
	err := isAnIdeaDirectory(directory)
	if err != nil {
		return nil, err
	}

	return &IdeaDirectory{directory}, nil
}

// Returned if InitIdeaDirectory is called on a directory
// that has already been initialized
var ErrInitOnExistingIdeaDirectory = errors.New("init on existing idea directory")

type ideaDirectoryInitialized struct {
	dir     string
	changes []git.CommitableChange
	msg     string
}

func (i ideaDirectoryInitialized) WorkingDirectory() string {
	return i.dir
}

func (i ideaDirectoryInitialized) Changes() []git.CommitableChange {
	return i.changes
}

func (i ideaDirectoryInitialized) CommitMsg() string {
	return i.msg
}

// Check that the directory is empty
// and if it is then it initializes an empty
// idea directory.
func InitIdeaDirectory(directory string) (*IdeaDirectory, git.Commitable, error) {
	err := isAnIdeaDirectory(directory)
	if err == nil {
		return nil, nil, ErrInitOnExistingIdeaDirectory
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

	return &IdeaDirectory{directory}, ideaDirectoryInitialized{
		directory,
		[]git.CommitableChange{
			git.ChangedFile("nextid"),
			git.ChangedFile("active"),
		},
		"idea directory initialized",
	}, nil
}

// Saves an idea to the idea directory and
// returns a commitable containing all changes.
// If the idea does not have an id it will be assigned one.
// If the idea does have an id it will be updated.
func (d IdeaDirectory) SaveIdea(idea *Idea) (git.Commitable, error) {
	return nil, nil
}

var ErrIdeaExists = errors.New("cannot save a new idea because it already exists")

// Saves an idea that doesn't have an id to the directory and
// returns a commitable containing all changes.
// If the idea is already assigned an id this method will
// return ErrIdeaExists
func (d IdeaDirectory) SaveNewIdea(idea *Idea) (git.Commitable, error) {
	return d.saveNewIdea(idea)
}

// Does not check if the idea has an id
func (d IdeaDirectory) saveNewIdea(idea *Idea) (git.Commitable, error) {
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
func (d IdeaDirectory) UpdateIdea(idea Idea) (git.Commitable, error) {
	return nil, nil
}
