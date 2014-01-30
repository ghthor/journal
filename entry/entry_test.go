package entry

import (
	"bufio"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func init() {
	var err error
	if _, err = os.Stat("_test/"); os.IsNotExist(err) {
		err = os.Mkdir("_test/", 0755)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func DescribeAnEntry(c gospec.Context) {
	td, err := ioutil.TempDir("_test", "entry_")
	c.Assume(err, IsNil)

	ne := New(td)

	c.Specify("an entry", func() {
		c.Specify("can be opened", func() {
			t := time.Date(2006, time.January, 1, 1, 0, 0, 0, time.UTC)

			c.Specify("at a specific time", func() {
				oe, err := ne.Open(t, nil)
				c.Assume(err, IsNil)
				c.Expect(oe.OpenedAt(), Equals, t)
			})

			c.Specify("with a list of ideas", func() {
				ideas := []idea.Idea{{
					Name:   "Active Idea",
					Status: idea.IS_Active,
					Body:   "Some text\n",
				}, {
					Name:   "Another Idea",
					Status: idea.IS_Active,
					Body:   "Some other text\n",
				}}

				oe, err := ne.Open(t, ideas)
				c.Assume(err, IsNil)
				actualIdeas, err := oe.Ideas()
				c.Assume(err, IsNil)
				c.Expect(len(actualIdeas), Equals, len(ideas))
				for i, idea := range actualIdeas {
					c.Expect(idea, Equals, ideas[i])
				}
			})
		})
		openedAt := time.Date(2006, time.January, 1, 1, 0, 0, 0, time.UTC)
		closedAt := time.Date(2006, time.January, 1, 1, 10, 0, 0, time.UTC)

		ideas := []idea.Idea{{
			Name:   "Active Idea",
			Status: idea.IS_Active,
			Body:   "Some text\n",
		}, {
			Name:   "Another Idea",
			Status: idea.IS_Active,
			Body:   "Some other text\n",
		}}

		oe, err := ne.Open(openedAt, ideas)
		c.Assume(err, IsNil)
		c.Assume(oe.OpenedAt(), Equals, openedAt)

		filename := filepath.Join(td, openedAt.Format(filenameLayout))

		c.Specify("that is open", func() {
			c.Specify("is a file", func() {
				_, err := os.Stat(filename)
				c.Expect(os.IsNotExist(err), IsFalse)

				actualBytes, err := ioutil.ReadFile(filename)
				c.Expect(err, IsNil)
				c.Expect(string(actualBytes), Equals,
					`Sun Jan  1 01:00:00 UTC 2006

#~ Title(will be used as commit message)
TODO Make this some random quote or something stupid

## [active] Active Idea
Some text

## [active] Another Idea
Some other text
`)
			})

			f, err := os.OpenFile(filename, os.O_RDONLY, 0600)
			c.Assume(err, IsNil)
			defer f.Close()

			c.Specify("will have the time opened as the first line of the entry", func() {
				scanner := bufio.NewScanner(f)
				c.Assume(scanner.Scan(), IsTrue)
				c.Expect(scanner.Text(), Equals, openedAt.Format(time.UnixDate))
			})

			c.Specify("can scan the entry for a list of ideas", func() {
				fa, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0600)
				c.Assume(err, IsNil)
				defer func() { c.Assume(fa.Close(), IsNil) }()

				_, err = fa.WriteString(
					`## [inactive] A Third Idea
Sir Idea, The Third
`)
				ideas = append(ideas, idea.Idea{
					Name:   "A Third Idea",
					Status: idea.IS_Inactive,
					Body:   "Sir Idea, The Third\n",
				})
				c.Assume(err, IsNil)

				actualIdeas, err := oe.Ideas()
				c.Assume(err, IsNil)
				c.Expect(len(actualIdeas), Equals, len(ideas))
				for i, actualIdea := range actualIdeas {
					c.Expect(actualIdea, Equals, ideas[i])
				}
			})

			c.Specify("will have a list of ideas appended to the entry", func() {
				actualIdeas, err := oe.Ideas()
				c.Assume(err, IsNil)
				c.Expect(len(actualIdeas), Equals, len(ideas))
				for i, actualIdea := range actualIdeas {
					c.Expect(actualIdea, Equals, ideas[i])
				}
			})

			c.Specify("can be editted by a text editor", func() {
				sed, err := exec.LookPath("sed")
				c.Assume(err, IsNil)

				editCmd := exec.Command(sed, "-i", "s_active_inactive_", filename)
				_, err = oe.Edit(editCmd)
				c.Expect(err, IsNil)

				// Re-Open the file
				c.Assume(f.Close(), IsNil)
				f, err = os.OpenFile(filename, os.O_RDONLY, 0600)
				c.Assume(err, IsNil)

				// Check the Edit's went through
				scanner := idea.NewIdeaScanner(f)
				for i := 0; i < len(ideas); i++ {
					c.Assume(scanner.Scan(), IsTrue)
					c.Expect(scanner.Idea().Status, Equals, idea.IS_Inactive)
				}
			})

			c.Specify("can be closed", func() {
				_, err = oe.Close(closedAt)
				c.Expect(err, IsNil)
			})

			c.Specify("cannot be closed without a commit msg", func() {
				err := ioutil.WriteFile(filename, []byte(
					`
A file without a commit message
`), 0600)
				c.Assume(err, IsNil)

				_, err = oe.Close(closedAt)
				c.Expect(err, Not(IsNil))
				c.Expect(err, Equals, ErrNoCommitMsg)
			})
		})

		c.Specify("that is closed", func() {
			ce, err := oe.Close(closedAt)
			c.Assume(err, IsNil)

			c.Specify("will have all ideas removed from the entry", func() {
				actualBytes, err := ioutil.ReadFile(filename)
				c.Assume(err, IsNil)

				c.Expect(string(actualBytes), Equals,
					`Sun Jan  1 01:00:00 UTC 2006

#~ Title(will be used as commit message)
TODO Make this some random quote or something stupid

Sun Jan  1 01:10:00 UTC 2006
`)
			})

			c.Specify("will have the time closed as the last line of the entry", func() {
				actualBytes, err := ioutil.ReadFile(filename)
				c.Assume(err, IsNil)

				c.Expect(string(actualBytes), Equals,
					`Sun Jan  1 01:00:00 UTC 2006

#~ Title(will be used as commit message)
TODO Make this some random quote or something stupid

Sun Jan  1 01:10:00 UTC 2006
`)
			})

			c.Specify("can be commited to the git repository", func() {
				commitable, isCommitable := ce.(git.Commitable)
				c.Expect(isCommitable, IsTrue)

				files, err := commitable.FilesToAdd()
				c.Assume(err, IsNil)
				c.Expect(len(files), Equals, 1)
				c.Expect(files[0], Equals, filename)

				commitMsg, err := commitable.CommitMsg()
				c.Assume(err, IsNil)
				c.Expect(commitMsg, Equals, "ENTRY - Title(will be used as commit message)")
			})
		})
	})
}
