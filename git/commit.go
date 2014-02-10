// A wrapper using "os/exec" for executing git commands
package git

type CommitableChange interface {
	Filepath() string
}

type Commitable interface {
	WorkingDirectory() string
	Changes() []CommitableChange
	CommitMsg() string
}

// An convenient implementation of the CommitableChange interface
type ChangedFile string

func (filename ChangedFile) Filepath() string { return string(filename) }

// An convenient implementation of the Commitable interface
type Changes struct {
	// The `git` working directory
	Dir string
	Msg string

	changes []CommitableChange
}

func NewChangesIn(workingDirectory string) *Changes {
	return &Changes{Dir: workingDirectory}
}

func (c *Changes) Add(change CommitableChange) {
	c.changes = append(c.changes, change)
}

func (c *Changes) Commit() error {
	return Commit(c)
}

// implement Commitable Interface

func (c Changes) WorkingDirectory() string    { return c.Dir }
func (c Changes) Changes() []CommitableChange { return c.changes }
func (c Changes) CommitMsg() string           { return c.Msg }

// Execute `git add` for all Changes()'s
// then execute `git commit` with CommitMsg()
func Commit(c Commitable) error {
	d := c.WorkingDirectory()

	for _, change := range c.Changes() {
		err := AddFilepath(d, change.Filepath())
		if err != nil {
			return err
		}
	}

	return CommitWithMessage(d, c.CommitMsg())
}
