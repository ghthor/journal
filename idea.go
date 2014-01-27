package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

const (
	IS_Active    = "active"
	IS_Inactive  = "inactive"
	IS_Completed = "completed"
)

type Idea struct {
	Name   string
	Status string
	Body   string
}

const IdeaTmpl = "## [{{.Status}}] {{.Name}}\n{{.Body}}"

type IdeaScanner struct {
	scanner *bufio.Scanner

	nextHeader []byte

	lastIdea  *Idea
	lastError error
}

func NewIdeaScanner(r io.Reader) *IdeaScanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanLines)
	return &IdeaScanner{scanner, nil, nil, nil}
}

// A ScanLines implementation that leaves the '\n' character in the token
// ripped from bufio.ScanLines
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

	// Parse the Status, Name, Body
	var prelimiter, status, startName string

	numScanned, err := fmt.Fscan(bytes.NewReader(sbuf.Bytes()), &prelimiter, &status, &startName)
	if err != nil {
		s.lastError = err
		return false
	}

	if numScanned != 3 {
		s.lastError = errors.New("unkown idea header format")
		return false
	}

	// Status
	// Strip "[" and "]"
	status = status[1 : len(status)-1]

	pbuf := bufio.NewReader(bytes.NewReader(sbuf.Bytes()))

	// Grab the complete header line as a string
	lineBytes, err := pbuf.ReadBytes('\n')
	if err != nil {
		s.lastError = err
		return false
	}
	line := string(lineBytes)

	// Name
	// Strip beginning of line and the '\n' byte
	name := line[strings.Index(line, startName) : len(line)-1]

	// Body
	bodyBytes, err := ioutil.ReadAll(pbuf)
	if err != nil {
		s.lastError = err
		return false
	}

	s.lastIdea = &Idea{
		Status: status,
		// Strip beginning of line and the '\n' byte
		Name: name,
		Body: string(bodyBytes),
	}

	return true
}

// Returns the last discovered Idea
func (s *IdeaScanner) Idea() *Idea {
	return s.lastIdea
}

func (s *IdeaScanner) Err() error {
	return s.lastError
}
