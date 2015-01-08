package fix

import (
	"os"
	"path/filepath"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
	"github.com/ghthor/journal/git/gittest"
)

func DescribeAFixableJournal(c gospec.Context) {
	c.Specify("a fixed journal is a", func() {
		// TODO this is a hack to create a fixed journal repo
		d, _, err := newCase0("case_current_spec")
		c.Assume(err, IsNil)

		_, err = Fix(d)
		c.Assume(err, IsNil)

		c.Specify("directory", func() {
			// c.Expect(d, isa, directory)
			c.Specify("inside a git repository", func() {
				// c.Expect(d, isa, gitrepo)
			})

			c.Specify("containing", func() {
				c.Specify("an entry directory", func() {
					// c.Expect(d, contains, "entry/")
					c.Specify("with all entries using the current entry format", func() {
						// c.Expect(d.entries, areCurrentFormat)
					})
				})

				c.Specify("an idea directory store", func() {
					// c.Expect(d.ideas, exists)
					// c.Expect(d.ideas, isEdittable)
				})
			})
		})
	})

	c.Specify("a fixable journal is a", func() {
		d, _, err := newCase0("case_0_is_fixable")
		c.Assume(err, IsNil)

		c.Specify("directory", func() {
			c.Specify("inside a git repository", func() {
				c.Assume(d, gittest.IsAGitRepository)
				c.Assume(git.IsClean(d), IsNil)

				// Case 0
				c.Specify("that contains entries", func() {
					needsFixed, err := NeedsFixed(d)
					c.Expect(needsFixed, IsTrue)

					_, err = Fix(d)
					c.Expect(err, IsNil)

					needsFixed, err = NeedsFixed(d)
					c.Expect(needsFixed, IsFalse)
				})
			})
		})
	})
}
