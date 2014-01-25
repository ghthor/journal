package main

import (
	"bufio"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
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
			testFilename := path.Join(jd, "test_file")
			f, err := os.OpenFile(testFilename, os.O_CREATE|os.O_RDWR, 0600)
			c.Assume(err, IsNil)

			fbuf := bufio.NewWriter(f)
			_, err = fbuf.WriteString("Some data that shouldn't be modified\n")
			c.Assume(err, IsNil)
			c.Assume(fbuf.Flush(), IsNil)
			c.Assume(f.Sync(), IsNil)

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
			c.Assume(f.Close(), IsNil)

			// Wait for watchEntry to complete it's task
			completedTime := <-timeCh

			f, err = os.OpenFile(testFilename, os.O_RDONLY, 0600)
			c.Assume(err, IsNil)
			defer f.Close()

			fscan := bufio.NewScanner(f)
			c.Expect(fscan.Scan(), IsTrue)
			c.Expect(fscan.Text(), Equals, "Some data that shouldn't be modified")
			c.Expect(fscan.Scan(), IsTrue)
			c.Expect(fscan.Text(), Equals, completedTime.Format(time.UnixDate))
			c.Expect(fscan.Scan(), IsFalse)
		})
	})
}
