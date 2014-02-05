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
	"io"
	"io/ioutil"
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
