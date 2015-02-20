package new

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

		c.Specify("will append the current time after editting is completed", func() {
			cmd := NewCmd(nil)
			cmd.SetWd(journalDir)

			// Mock time to control the filename and openedAt/closedAt times stored in the entry
			openedAt := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
			closedAt := time.Date(2015, 1, 2, 0, 0, 0, 0, time.UTC)

			// This is a roundabout mock from hell...but it works...
			// TODO figure more elegant mock for this
			var nowFn func() time.Time
			nowFn = func() time.Time {
				// Mutate during first call to return ClosedAt time
				nowFn = func() time.Time { return closedAt }
				// OpenedAt time
				return openedAt
			}
			cmd.Now = func() time.Time { return nowFn() }

			entryFilename := openedAt.Format(entry.FilenameLayout)

			// Mocked editor that does nothing
			cmd.EditorProcess = mockEditor{
				start: func() {},
				wait:  func() {},
			}

			// Run `journal new` with mocked EditorProcess and Now functions
			c.Assume(cmd.Exec(nil), IsNil)

			// Entry will have closing time appended
			f, err := os.OpenFile(filepath.Join(journalDir, "entry", entryFilename), os.O_RDONLY, 0600)
			c.Assume(err, IsNil)
			defer f.Close()

			scanner := bufio.NewScanner(f)
			scanner.Split(bufio.ScanLines)

			var prevLine string
			for scanner.Scan() {
				c.Assume(scanner.Err(), IsNil)
				prevLine = scanner.Text()
			}

			t, err := time.Parse(time.UnixDate, prevLine)
			c.Assume(err, IsNil)
			c.Expect(t, Equals, closedAt)
		})

		c.Specify("will commit the entry to the git repository", func() {
			cmd := NewCmd(nil)
			cmd.SetWd(journalDir)

			// Mock time to control the filename and openedAt/closedAt times stored in the entry
			openedAt := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
			cmd.Now = func() time.Time { return openedAt }

			// Mocked editor that does nothing
			cmd.EditorProcess = mockEditor{
				start: func() {},
				wait:  func() {},
			}

			// Run `journal new` with mocked EditorProcess and Now functions
			c.Assume(cmd.Exec(nil), IsNil)

			// Entry will be shown in the git repository
			c.Expect(git.IsClean(journalDir), IsNil)

			// Save test output dir for manual inspection
			// c.Assume(exec.Command("cp", "-r", journalDir, filepath.Join("/tmp", "new_cmd_git_commit")).Run(), IsNil)

			lastCommitBytes, err := git.Command(journalDir, "show", "--pretty=format:%T").Output()
			c.Assume(err, IsNil)
			c.Expect(string(lastCommitBytes), Equals, `2d0f3cbd0e6d8409c9bb767b7bcc09fb569eaa06
diff --git a/entry/2015-01-01-0000-UTC b/entry/2015-01-01-0000-UTC
new file mode 100644
index 0000000..c85666f
--- /dev/null
+++ b/entry/2015-01-01-0000-UTC
@@ -0,0 +1,6 @@
+Thu Jan  1 00:00:00 UTC 2015
+
+# Title(will be used as commit message)
+TODO Make this some random quote or something stupid
+
+Thu Jan  1 00:00:00 UTC 2015
`)
			hashAndTitleBytes, err := git.Command(journalDir, "show", "-s", "--format=%s").Output()
			c.Assume(err, IsNil)
			c.Expect(string(hashAndTitleBytes), Equals, "Title(will be used as commit message)\n")
		})

		c.Specify("will commit any modifications to the idea store", func() {
			cmd := NewCmd(nil)
			cmd.SetWd(journalDir)

			// Create an active idea
			store, err := idea.NewDirectoryStore(filepath.Join(journalDir, "idea"))
			c.Assume(err, IsNil)

			activeIdea := idea.Idea{
				Status: idea.IS_Active,
				Name:   "tset idea",
				Body:   "test idea body\n",
			}

			commitable, err := store.SaveIdea(&activeIdea)
			c.Assume(err, IsNil)
			c.Assume(git.Commit(commitable), IsNil)

			// Mock time to control the filename and openedAt/closedAt times stored in the entry
			openedAt := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
			cmd.Now = func() time.Time { return openedAt }

			entryFilename := openedAt.Format(entry.FilenameLayout)

			// Edit to fix the misspelled title in the idea
			editCmd := exec.Command("sed", "-i", "s_tset_test_", entryFilename)
			editCmd.Dir = filepath.Join(journalDir, "entry")
			cmd.EditorProcess = editCmd

			// Run `journal new` with mocked EditorProcess and Now functions
			c.Assume(cmd.Exec(nil), IsNil)

			c.Expect(git.IsClean(journalDir), IsNil)

			// Save test output dir for manual inspection
			// c.Assume(exec.Command("cp", "-r", journalDir, filepath.Join("/tmp", "new_cmd_git_commit")).Run(), IsNil)

			lastCommitBytes, err := git.Command(journalDir, "show", "--pretty=format:%T", "HEAD^").Output()
			c.Assume(err, IsNil)
			c.Expect(string(lastCommitBytes), Equals, `d6551c5250432aaa5244aa7767ae672cb316c1bb
diff --git a/idea/1 b/idea/1
index 83f5e84..0b22af3 100644
--- a/idea/1
+++ b/idea/1
@@ -1,2 +1,2 @@
-## [active] [1] tset idea
+## [active] [1] test idea
 test idea body
`)
			hashAndTitleBytes, err := git.Command(journalDir, "show", "-s", "--format=%s", "HEAD^").Output()
			c.Assume(err, IsNil)
			c.Expect(string(hashAndTitleBytes), Equals, "idea - updated - 1\n")

		})

		c.Specify("will fail", func() {
			cmd := NewCmd(nil)
			cmd.SetWd(journalDir)

			c.Specify("if the journal directory has a dirty git repository", func() {
				// Dirty the test journal
				c.Assume(exec.Command("touch", filepath.Join(journalDir, "makedirty")).Run(), IsNil)
				c.Assume(git.IsClean(journalDir), Not(IsNil))

				// Mocked editor that does nothing
				cmd.EditorProcess = mockEditor{
					start: func() {},
					wait:  func() {},
				}

				// Will fail with an error
				c.Expect(cmd.Exec(nil), Equals, ErrGitIsDirty)
			})
		})
	})

	c.Specify("the environment editor", func() {
		c.Specify("will be vim", func() {
			cmd, err := newEnvEditor("vim", "entryFilename")
			c.Assume(err, IsNil)

			c.Expect(filepath.Base(cmd.Args[0]), Equals, "vim")
			c.Expect(strings.Join(cmd.Args[1:], " "), Equals, "+set spell entryFilename")
		})

		c.Specify("will be emacs", func() {
			cmd, err := newEnvEditor("emacs", "entryFilename")
			c.Assume(err, IsNil)

			c.Expect(filepath.Base(cmd.Args[0]), Equals, "emacs")
			c.Expect(strings.Join(cmd.Args[1:], " "), Equals, "entryFilename")

			cmd, err = newEnvEditor("emacs -nw", "entryFilename")
			c.Assume(err, IsNil)

			c.Expect(filepath.Base(cmd.Args[0]), Equals, "emacs")
			c.Expect(strings.Join(cmd.Args[1:], " "), Equals, "-nw entryFilename")
		})
	})
}
