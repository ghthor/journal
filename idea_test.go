package main

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"strings"
)

func DescribeIdea(c gospec.Context) {
	c.Specify("an Idea", func() {
		const someData = `
Some other markdowned text.
Doesn't matter waht it is

## [active] An Idea
Some text explaining the idea.

And some more.

There isn't a delimiter.
So this next idea will be the delimiter.

## [active] Another Idea
And the delimiter for this idea will be the ClosedAt timestamp.
That is at the end of every entry.

The newline before the timestamp should be included.

Sun Jan 26 15:03:44 EST 2014
`

		c.Specify("will be discovered", func() {
			scanner := NewIdeaScanner(strings.NewReader(someData))

			c.Specify("and will include everything from the header to the next idea", func() {
				c.Expect(scanner.Scan(), IsTrue)
				c.Expect(scanner.Err(), IsNil)

				idea := scanner.Idea()
				c.Expect(idea.Name, Equals, "An Idea")
				c.Expect(idea.Status, Equals, IS_Active)
				c.Expect(idea.Body, Equals,
					`Some text explaining the idea.

And some more.

There isn't a delimiter.
So this next idea will be the delimiter.
`)
			})

			c.Specify("and will not include the date from the EOF", func() {
				// Drop the first idea
				c.Assume(scanner.Scan(), IsTrue)

				// Scan the second idea
				c.Expect(scanner.Scan(), IsTrue)
				c.Expect(scanner.Err(), IsNil)

				idea := scanner.Idea()
				c.Expect(idea.Name, Equals, "Another Idea")
				c.Expect(idea.Status, Equals, IS_Active)
				c.Expect(idea.Body, Equals,
					`And the delimiter for this idea will be the ClosedAt timestamp.
That is at the end of every entry.

The newline before the timestamp should be included.
`)
			})
		})
	})
}
