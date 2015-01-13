package new

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ghthor/journal/entry"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/idea"
	initialize "github.com/ghthor/journal/init"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type mockEditor struct {
	start, wait func()
}

func (m mockEditor) Start() error {
	m.start()
	return nil
}

func (m mockEditor) Wait() error {
	m.wait()
	return nil
}

func DescribeNewCmd(c gospec.Context) {
	c.Specify("the `new` command", func() {
		// Create a temporary journal
		journalDir, err := ioutil.TempDir("", "new_cmd_desc_")
		c.Assume(err, IsNil)
		defer func() {
			c.Assume(os.RemoveAll(journalDir), IsNil)
		}()

		commitable, err := initialize.Journal(journalDir)
		c.Assume(err, IsNil)
		c.Assume(git.Commit(commitable), IsNil)

		c.Specify("will include any active ideas in the entries body while editting", func() {
			// Create an active idea
			store, err := idea.NewDirectoryStore(filepath.Join(journalDir, "idea"))
			c.Assume(err, IsNil)

			activeIdea := idea.Idea{
				Status: idea.IS_Active,
				Name:   "test idea",
				Body:   "test idea body\n",
			}

			commitable, err := store.SaveIdea(&activeIdea)
			c.Assume(err, IsNil)
			c.Assume(git.Commit(commitable), IsNil)

			cmd := NewCmd(nil)
			cmd.SetWd(journalDir)

			cmd.Now = func() time.Time {
				return time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			entryFilename := cmd.Now().Format(entry.FilenameLayout)

			editorProcessHasStarted := make(chan bool)
			expectationsChecked := make(chan bool)
			execCompleted := make(chan bool)

			cmd.EditorProcess = mockEditor{
				start: func() {
					editorProcessHasStarted <- true
				},
				wait: func() {
					<-expectationsChecked
				},
			}

			// Run `journal new` with mocked EditorProcess and Now functions
			go func() {
				c.Assume(cmd.Exec(nil), IsNil)
				execCompleted <- true
			}()

			<-editorProcessHasStarted

			// Entry will have that active idea as it is being editted
			f, err := os.OpenFile(filepath.Join(journalDir, "entry", entryFilename), os.O_RDONLY, 0600)
			c.Assume(err, IsNil)
			defer f.Close()

			ideaScanner := idea.NewIdeaScanner(f)
			ideaScanner.Scan()
			c.Assume(ideaScanner.Err(), IsNil)

			idea := ideaScanner.Idea()
			c.Assume(idea, Not(IsNil))
			c.Expect(*idea, Equals, activeIdea)

			// sync execution back up
			expectationsChecked <- true
			<-execCompleted
		})

		c.Specify("will update the idea store with any modifications made during editting", func() {
			// Create an idea
			store, err := idea.NewDirectoryStore(filepath.Join(journalDir, "idea"))
			c.Assume(err, IsNil)

			activeIdea := idea.Idea{
				Status: idea.IS_Active,
				Name:   "test idea",
				Body:   "test idea body\n",
			}

			commitable, err := store.SaveIdea(&activeIdea)
			c.Assume(err, IsNil)
			c.Assume(git.Commit(commitable), IsNil)

			cmd := NewCmd(nil)
			cmd.SetWd(journalDir)

			cmd.Now = func() time.Time {
				return time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			entryFilename := cmd.Now().Format(entry.FilenameLayout)

			sedCmd := exec.Command("sed", "-i", "s_active_inactive_", entryFilename)
			sedCmd.Dir = filepath.Join(journalDir, "entry")

			cmd.EditorProcess = sedCmd

			c.Expect(cmd.Exec(nil), IsNil)

			// Modify the status to reflect what happened during the edit
			activeIdea.Status = idea.IS_Inactive

			// Idea in the IdeaStore will be updated if it was editted
			idea, err := store.IdeaById(activeIdea.Id)
			c.Assume(err, IsNil)
			c.Expect(idea, Equals, activeIdea)
		})

		c.Specify("will commit the entry to the git repository", func() {
			// Run `new`
			// Will succeed
			// Entry will be shown in the git repository

			c.Specify("and will commit any modifications to the idea store", func() {
				// Any modifed Ideas will also have commits
			})
		})

		c.Specify("will append the current time after editting is completed", func() {
			// Run `new`
			// Will succeed
			// Entry will have closing time appended
		})

		c.Specify("will fail", func() {
			// Dirty the test journal
			c.Specify("if the journal directory has a dirty git repository", func() {
				// Run `new`
				// Will fail with an error
			})
		})
	})
}
