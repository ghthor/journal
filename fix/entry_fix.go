package fix

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/ghthor/journal/idea"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

var entryFixes []entryFix

func init() {
	entryFixes = []entryFix{
		fixAddClosedAtTimestamp{},
		FixSplitCommitMessage{},
		FixCommitMessagePrefixWithTilde{},
		FixIdeasInBody{},
	}
}

type entryFix interface {
	// Parses io.Reader for the error that can be fixed
	CanFix(io.Reader) (bool, error)

	// Returns byte slice that has been fixed
	Execute(io.Reader) ([]byte, error)
}

// Parse the opened at timestamp and add 2 mins
// then append it to the end
type fixAddClosedAtTimestamp struct{}

func (fixAddClosedAtTimestamp) CanFix(r io.Reader) (bool, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return false, err
	}

	return !lastLineIsTimestamp(data), nil
}

func lastLineIsTimestamp(data []byte) bool {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Scan to the last line
	var prevLine string
	for scanner.Scan() {
		prevLine = scanner.Text()
	}
	_, err := time.Parse(time.UnixDate, prevLine)

	return err == nil
}

func (fixAddClosedAtTimestamp) Execute(r io.Reader) ([]byte, error) {
	// For adding 2 mins and making the closed at timestamp
	var openedAt time.Time

	// For storing the fixed output
	b := bytes.NewBuffer(make([]byte, 0, 1024))

	scanner := bufio.NewScanner(r)

	// Scan in the opened at timestamp
	if scanner.Scan() {
		if time, err := time.Parse(time.UnixDate, scanner.Text()); err != nil {
			// First line wasn't a timestamp
			return nil, err
		} else {
			openedAt = time
			// Print the timestamp to the fixed buffer
			if _, err := fmt.Fprintln(b, scanner.Text()); err != nil {
				return nil, err
			}
		}
	} else {
		return nil, errors.New("error parsing opened at timestamp")
	}

	// Scan and copy lines into the fixed buffer
	for scanner.Scan() {
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}

		if _, err := fmt.Fprintln(b, scanner.Text()); err != nil {
			return nil, err
		}
	}

	// Append the closed at timestamp to the end of the buffer
	closedAt := openedAt.Add(time.Minute * 2)
	if _, err := fmt.Fprintf(b, "\n%s\n", closedAt.Format(time.UnixDate)); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

/*
	Fix a commit message with the format

		#~ Commit Msg
		# Additional Msg

	By turning it into this

		# Commit Msg | Additional Msg

*/
type FixSplitCommitMessage struct{}

func (FixSplitCommitMessage) CanFix(r io.Reader) (bool, error) {
	return hasSplitCommitMsg(r), nil
}

func hasSplitCommitMsg(r io.Reader) bool {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#~ ") {
			if scanner.Scan() {
				if strings.HasPrefix(scanner.Text(), "# ") {
					return true
				}
			}
		}
	}
	return false
}

func (FixSplitCommitMessage) Execute(r io.Reader) ([]byte, error) {
	// For storing the fixed output
	b := bytes.NewBuffer(make([]byte, 0, 1024))

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#~ ") {
			// We've found line 1 of the split commit message
			part1 := strings.TrimPrefix(line, "#~ ")

			if scanner.Scan() {
				// Trim line 2 of the message and put part1 and part2 together
				part2 := strings.TrimPrefix(scanner.Text(), "# ")

				if _, err := fmt.Fprintf(b, "# %s | %s\n", part1, part2); err != nil {
					return nil, err
				}
			} else {
				// Maybe this should be a panic
				return nil, errors.New("attempt to fix split commit message that doesn't exist")
			}
			break
		}

		if _, err := fmt.Fprintln(b, line); err != nil {
			return nil, err
		}
	}

	// Copy the remaining bytes into the buffer
	for scanner.Scan() {
		if _, err := fmt.Fprintln(b, scanner.Text()); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

/*
	Fix a commit message line with the #~ format

		#~ Commit Msg

	to

		# Commit Msg

*/
type FixCommitMessagePrefixWithTilde struct{}

func (FixCommitMessagePrefixWithTilde) CanFix(r io.Reader) (bool, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#~ ") {
			if scanner.Scan() {
				// Make sure this isn't a split commit message
				if !strings.HasPrefix(scanner.Text(), "# ") {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (FixCommitMessagePrefixWithTilde) Execute(r io.Reader) ([]byte, error) {
	// For storing the fixed output
	b := bytes.NewBuffer(make([]byte, 0, 1024))

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#~ ") {
			// Trim out the ~ character and place in the buffer
			if _, err := fmt.Fprintln(b, strings.Replace(line, "#~", "#", 1)); err != nil {
				return nil, err
			}
			break
		}

		if _, err := fmt.Fprintln(b, line); err != nil {
			return nil, err
		}
	}

	// Copy the remaining bytes into the buffer
	for scanner.Scan() {
		if _, err := fmt.Fprintln(b, scanner.Text()); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

/*
	Remove any idea's from the entries body.
	We assume that the ideas have already been
	parsed and saved in an earlier fix step.
*/
type FixIdeasInBody struct{}

func (FixIdeasInBody) CanFix(r io.Reader) (bool, error) {
	return idea.NewIdeaScanner(r).Scan(), nil
}

func (FixIdeasInBody) Execute(r io.Reader) ([]byte, error) {
	// For storing the fixed output
	b := bytes.NewBuffer(make([]byte, 0, 1024))

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "## [") {
			break
		}

		if _, err := fmt.Fprintln(b, scanner.Text()); err != nil {
			return nil, err
		}
	}

	var timestampLine string

	// Look for the timestamp
	for scanner.Scan() {
		if _, err := time.Parse(time.UnixDate, scanner.Text()); err == nil {
			timestampLine = scanner.Text()
		}
	}

	// Trim what's in the buffer and append the timestamp line
	b = bytes.NewBuffer(bytes.TrimSpace(b.Bytes()))
	if _, err := fmt.Fprintf(b, "\n\n%s\n", timestampLine); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
