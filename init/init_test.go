package init_test

import (
	"github.com/ghthor/gospec"
)

func DescribeInit(c gospec.Context) {
	c.Specify("a journal", func() {
		c.Specify("that is initialized", func() {
			c.Specify("is a git repository", func() {
				c.Specify("that contains", func() {
					c.Specify("an entry directory", func() {
						c.Specify("that can have entries", func() {
						})
					})

					c.Specify("an idea directory store", func() {
						c.Specify("that can have ideas", func() {
						})
					})
				})
			})
		})

		c.Specify("that can be initialized", func() {
			c.Specify("is an empty directory", func() {
				c.Specify("inside a git repository", func() {
				})

				c.Specify("NOT inside a git repository", func() {
				})
			})
		})

		c.Specify("that can NOT be initialized", func() {
			c.Specify("is a directory", func() {
				c.Specify("that already contains", func() {
					c.Specify("an entry directory", func() {
					})

					c.Specify("an idea directory store", func() {
					})
				})
			})
		})
	})
}
