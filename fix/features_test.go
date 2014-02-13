package fix

import (
	"github.com/ghthor/gospec"
)

func DescribeAFixableJournal(c gospec.Context) {
	c.Specify("a fixed journal is a", func() {
		c.Specify("directory", func() {
			c.Specify("inside a git repository", func() {
			})

			c.Specify("containing", func() {
				c.Specify("an entry directory", func() {
				})

				c.Specify("an idea directory store", func() {
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
