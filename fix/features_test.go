package fix

import (
	"github.com/ghthor/gospec"
)

func DescribeAFixableJournal(c gospec.Context) {
	c.Specify("a fixed journal is a", func() {
		// d, _, _ := newCaseCurrent("case_current_spec")
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
		c.Specify("directory", func() {
			c.Specify("inside a git repository", func() {
				c.Specify("that contains entries", func() {
				})

				c.Specify("that contains no entries", func() {
				})

				c.Specify("that is empty", func() {
				})
			})
		})
	})
}
