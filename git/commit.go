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
	wd      string
	changes []CommitableChange
	msg     string
}

func NewChangesIn(workingDirectory string) *Changes {
	return &Changes{wd: workingDirectory}
}

func (c *Changes) Add(change CommitableChange) {
	c.changes = append(c.changes, change)
}

func (c *Changes) Commit(msg string) error {
	c.msg = msg
	return Commit(c)
}

// implement Commitable Interface

func (c Changes) WorkingDirectory() string    { return c.wd }
func (c Changes) Changes() []CommitableChange { return c.changes }
func (c Changes) CommitMsg() string           { return c.msg }

// Execute a `git commit` with some commitable changes
func Commit(c Commitable) error {
	d := c.WorkingDirectory()

	for _, change := range c.Changes() {
		err := GitAdd(d, change.Filepath())
		if err != nil {
			return err
		}
	}

	return GitCommit(d, c.CommitMsg())
}
