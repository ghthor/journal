package git

type Commitable interface {
	// Slice of Filepaths to `git add`
	FilesToAdd() ([]string, error)
	// Commit message for `git commit`
	CommitMsg() (string, error)
}
