package main

import (
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"io/ioutil"
	"os"
	"path"
	"time"
)

func DescribeWatchCmd(c gospec.Context) {
	jd, err := tmpGitRepository("journal_test")
	c.Assume(err, IsNil)

	defer func() {
		err := os.RemoveAll(jd)
		c.Assume(err, IsNil)
	}()

	c.Specify("the `watch` command", func() {
		c.Specify("will append the time the entry was completed", func() {
			fileData := "Some data that shouldn't be modified"
			testFilename := path.Join(jd, "test_file")

			c.Assume(ioutil.WriteFile(testFilename, []byte(fileData+"\n"), 0600), IsNil)

			statusCh := make(chan string)
			timeCh := make(chan time.Time)

			go func() {
				t, err := watchEntry(statusCh, &Command{}, testFilename)
				c.Assume(err, IsNil)

				timeCh <- t
			}()

			// Wait for watchEntry to be watching "test_file" via inotify
			c.Expect(<-statusCh, Equals, "watching")

			// Trigger watchEntry via inotify watcher
			f, err := os.Open(testFilename)
			c.Assume(err, IsNil)
			c.Assume(f.Close(), IsNil)

			// Wait for watchEntry to complete it's task
			completedTime := <-timeCh

			actualData, err := ioutil.ReadFile(testFilename)
			c.Assume(err, IsNil)
			c.Expect(string(actualData), Equals, fmt.Sprintf("%s\n\n%s\n", fileData, completedTime.Format(time.UnixDate)))
		})
	})
}
