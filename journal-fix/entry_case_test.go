package main

import (
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const entry_case_0 = `Sun Jan 26 14:50:50 EST 2014

# Commit Msg | A little extra
Body
`

const entry_case_1 = `Sun Jan 26 14:50:50 EST 2014

# Commit Msg | A little extra
Body

Sun Jan 26 14:52:50 EST 2014
`
const entry_case_2 = `Sun Jan 26 14:50:50 EST 2014

#~ Commit Msg
# A little extra
Body

Sun Jan 26 14:52:50 EST 2014
`

const entry_case_3 = `Sun Jan 26 14:50:50 EST 2014

#~ Commit Msg | A little extra
Body

Sun Jan 26 14:52:50 EST 2014
`

const entry_case_4 = `Sun Jan 26 14:50:50 EST 2014

#~ Commit Msg | A little extra
Body

## [status] Idea
Idea Body

Sun Jan 26 14:52:50 EST 2014
`

const entry_case_current = entry_case_1

func DescribeEntry(c gospec.Context) {
	entryCasesData := []string{
		entry_case_0,
		entry_case_1,
		entry_case_2,
		entry_case_3,
		entry_case_4,
		entry_case_current,
	}
	entryCases := make([]Entry, 0, len(entryCasesData))

	for _, data := range entryCasesData {
		entryCase, err := NewEntry(strings.NewReader(data))
		c.Assume(err, IsNil)

		entryCases = append(entryCases, entryCase)
	}

	c.Specify("an entry case", func() {
		c.Specify("can be read", func() {
			c.Specify("from an io.Reader", func() {
				for _, data := range entryCasesData {
					entryCase, err := NewEntry(strings.NewReader(data))
					c.Expect(err, IsNil)
					c.Expect(string(entryCase.Bytes()), Equals, string(data))
				}
			})

			c.Specify("from a file", func() {
				d, err := ioutil.TempDir("_test", "entry_can_be_read_from_file_")
				c.Assume(err, IsNil)

				for i, data := range entryCasesData {
					filename := filepath.Join(d, fmt.Sprintf("case_%d", i))
					c.Assume(ioutil.WriteFile(filename, []byte(data), 0600), IsNil)

					entryCase, err := NewEntryFromFile(filename)
					c.Expect(err, IsNil)

					actualData, err := ioutil.ReadAll(entryCase.NewReader())
					c.Assume(err, IsNil)
					c.Expect(string(actualData), Equals, string(data))
				}
			})
		})

		c.Specify("can be fixed", func() {
			entriesFixed := make([]Entry, 0, len(entryCases))

			for _, entryCase := range entryCases {
				entriesFixed = append(entriesFixed, entryCase.FixedEntry())
			}

			c.Specify("by returning an entry case for the current standard", func() {
				for _, fixedEntry := range entriesFixed {
					actualData, err := ioutil.ReadAll(fixedEntry.NewReader())
					c.Assume(err, IsNil)

					c.Expect(string(actualData), Equals, string("the expected output"))
				}
			})
		})

		c.Specify("can be written", func() {
			for i, entryCase := range entryCases {
				actualData, err := ioutil.ReadAll(entryCase.NewReader())
				c.Assume(err, IsNil)
				c.Expect(string(actualData), Equals, entryCasesData[i])
			}
		})
	})
}
