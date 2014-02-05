// Package Idea implements io.Reader's and a Scanner for
// reading and writing the raw text Idea format.
// It also provides a way to create and update Idea's
// stored in text files within a directory.
package idea

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/ghthor/journal/git"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const (
	// Valid Idea.Status values
	IS_Active    = "active"
	IS_Inactive  = "inactive"
	IS_Completed = "completed"
)

type Idea struct {
	Status string
	Id     uint
	Name   string
	Body   string
}

var ideaTmpl = template.Must(template.New("idea").Parse(
	`## [{{.Status}}] {{if .Id}}[{{.Id}}] {{end}}{{.Name}}
{{.Body}}`))

// An implementation of io.Reader for Idea
type IdeaReader struct {
	buf io.Reader
}

// Create an io.Reader with idea
func NewIdeaReader(idea Idea) (*IdeaReader, error) {
	b := bytes.NewBuffer(make([]byte, 0, 1024))

	err := ideaTmpl.Execute(b, idea)
	if err != nil {
		return nil, err
	}

	return &IdeaReader{b}, nil
}

func (r *IdeaReader) Read(b []byte) (n int, err error) {
	return r.buf.Read(b)
}

// Used to scan Idea's from an io.Reader
type IdeaScanner struct {
	scanner *bufio.Scanner

	nextHeader []byte

	lastIdea  *Idea
	lastError error
}

// Create an IdeaScanner using an io.Reader
func NewIdeaScanner(r io.Reader) *IdeaScanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanLines)
	return &IdeaScanner{scanner, nil, nil, nil}
}

// A ScanLines implementation that leaves the '\n' character in the token.
// Ripped from bufio.ScanLines
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func parseHeader(raw string) (status string, id uint, name string, err error) {
	_, err = fmt.Fscanf(strings.NewReader(raw), "## %s [%d] %s", &status, &id, &name)
	if err != nil {
		switch err.Error() {
		default:
			return
		case "expected integer":
			_, err = fmt.Fscanf(strings.NewReader(raw), "## %s [] %s", &status, &name)
		case "input does not match format":
			_, err = fmt.Fscanf(strings.NewReader(raw), "## %s %s", &status, &name)
		}

		if err != nil {
			return
		}
	}

	if status[0] != '[' && status[len(status)-1] != ']' {
		err = errors.New("invalid idea header: status must be wrapped w/ []")
		return
	}

	status = strings.Trim(status, "[]")
	name = raw[strings.Index(raw, name):]
	name = strings.TrimRight(name, "\n")

	return
}

// Scans through the io.Reader for the next Idea
func (s *IdeaScanner) Scan() bool {
	sbuf := bytes.NewBuffer(make([]byte, 0, 4096))

	// Scan the Header
	if s.nextHeader != nil {
		_, err := sbuf.Write(s.nextHeader)
		if err != nil {
			s.lastError = err
			return false
		}
	} else {
		for s.scanner.Scan() {
			line := s.scanner.Text()

			if i := strings.Index(line, "## ["); i >= 0 {
				_, err := sbuf.Write(s.scanner.Bytes())
				if err != nil {
					s.lastError = err
					return false
				}

				goto scanBody
			}
		}

		// Scanned all the lines and didn't find a header
		s.lastError = s.scanner.Err()
		return false
	}

scanBody:
	s.nextHeader = nil

	for s.scanner.Scan() {
		line := s.scanner.Text()

		// Look for the start of another Idea
		if i := strings.Index(line, "## ["); i >= 0 {
			s.nextHeader = s.scanner.Bytes()
			break
		}

		// Look for the ClosedAt timestamp
		if _, err := time.Parse(time.UnixDate, line[:len(line)-1]); err == nil {
			break
		}

		// Add this line to the body
		if _, err := sbuf.Write(s.scanner.Bytes()); err != nil {
			s.lastError = err
			return false
		}
	}

	pbuf := bufio.NewReader(bytes.NewReader(sbuf.Bytes()))

	// Grab the complete header line as a string
	lineBytes, err := pbuf.ReadBytes('\n')
	if err != nil {
		s.lastError = err
		return false
	}
	line := string(lineBytes)

	// Parse the Status, Id, Name
	status, id, name, err := parseHeader(line)
	if err != nil {
		s.lastError = err
		return false
	}

	// Parse the Body
	bodyBytes, err := ioutil.ReadAll(pbuf)
	if err != nil {
		s.lastError = err
		return false
	}

	s.lastIdea = &Idea{
		Status: status,
		Id:     id,
		Name:   name,
		Body:   string(bytes.TrimSpace(bodyBytes)) + "\n",
	}

	return true
}

// Returns the last discovered Idea
func (s *IdeaScanner) Idea() *Idea {
	return s.lastIdea
}

// Returns the last error
func (s *IdeaScanner) Err() error {
	return s.lastError
}

// Used to manage idea storage in a directory
type IdeaDirectory struct {
	directory string
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
	changes := git.NewChangesIn(d.directory)

	// Retrieve nextid
	data, err := ioutil.ReadFile(filepath.Join(d.directory, "nextid"))
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

	err = ioutil.WriteFile(filepath.Join(d.directory, "nextid"), []byte(fmt.Sprintf("%d\n", nextId)), 0600)
	if err != nil {
		return nil, err
	}
	changes.Add(git.ChangedFile("nextid"))

	// write to file
	r, err := NewIdeaReader(*idea)
	if err != nil {
		return nil, err
	}

	ideaFile, err := os.OpenFile(filepath.Join(d.directory, fmt.Sprint(idea.Id)), os.O_CREATE|os.O_WRONLY, 0600)
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
		activeIndexFile, err := os.OpenFile(filepath.Join(d.directory, "active"), os.O_APPEND|os.O_WRONLY, 0600)
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
