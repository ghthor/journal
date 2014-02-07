package fix

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
)

func lastLineIsTimestamp(data []byte) bool {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Scan to the last line
	for scanner.Scan() {
	}
	_, err := time.Parse(time.UnixDate, scanner.Text())

	return err == nil
}

// Parse the opened at timestamp and add 2 mins
// then append it to the end
type AddClosedAtTimestamp struct{}

// Parse the opened at timestamp and add 2 mins
// then append it to the end
func (AddClosedAtTimestamp) Fix(r io.Reader) ([]byte, error) {
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
