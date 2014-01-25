package main

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"io/ioutil"
	"os"
	"path"
)

func DescribeNewCmd(c gospec.Context) {
	jd, err := tmpGitRepository("journal_test")
	c.Assume(err, IsNil)

	defer func() {
		err := os.RemoveAll(jd)
		c.Assume(err, IsNil)
	}()

	c.Specify("the `new` command", func() {
		c.Specify("will fail", func() {
			c.Specify("if the journal directory has a dirty git repository", func() {
				c.Assume(ioutil.WriteFile(path.Join(jd, "dirty"), []byte("some data"), os.FileMode(0600)), IsNil)
				err := newEntry(jd, false, &Command{})
				c.Expect(err, Not(IsNil))
			})
		})
	})
}
