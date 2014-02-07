package main

import (
	"github.com/ghthor/gospec"
)

func DescribeEntryCase(c gospec.Context) {
	c.Specify("an entry case", func() {
		c.Specify("can be read", func() {
			c.Specify("from an io.Reader", func() {
			})

			c.Specify("from a file", func() {
			})
		})

		c.Specify("can be fixed", func() {
			c.Specify("by returning an entry case for the current standard", func() {
			})
		})

		c.Specify("can be written", func() {
		})
	})
}
