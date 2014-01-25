package main

import (
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"io/ioutil"
	"os"
	"path"
	"text/template"
)

func DescribeNewCmd(c gospec.Context) {
	jd, err := tmpGitRepository("journal_test")
	c.Assume(err, IsNil)

	defer func() {
		err := os.RemoveAll(jd)
		c.Assume(err, IsNil)
	}()

	c.Specify("the `new` command", func() {
		c.Specify("will append the current time after editting is completed", func() {
			const fileData = "Some data that shouldn't be modified\n"
			j, err := newEntry(jd, template.Must(template.New(jd).Parse(fileData)), nil, &Command{})
			c.Assume(err, IsNil)

			actualData, err := ioutil.ReadFile(path.Join(jd, j.Filename))
			c.Assume(err, IsNil)

			c.Expect(string(actualData), Equals, fmt.Sprintf("%s\n%s\n", fileData, j.ClosedAt))
		})

		c.Specify("will fail", func() {
			c.Specify("if the journal directory has a dirty git repository", func() {
				c.Assume(ioutil.WriteFile(path.Join(jd, "dirty"), []byte("some data"), os.FileMode(0600)), IsNil)
				_, err := newEntry(jd, entryTmpl, nil, &Command{})
				c.Expect(err, Not(IsNil))
			})
		})
	})
}
