package new

import (
	"github.com/ghthor/gospec"
)

func DescribeNewCmd(c gospec.Context) {
	c.Specify("the `new` command", func() {
		// Create a test journal

		c.Specify("will include any active ideas", func() {
			// Create an active idea
			// Run `new`
			// Will succeed
			// Entry will have that active idea as it is being editted

			c.Specify("and will update the idea store with any modifications", func() {
				// Idea can be editted from within the entry
				// Idea in the IdeaStore will be updated if it was editted
			})
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
