package idea

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"strings"
)

func DescribeIdea(c gospec.Context) {
	c.Specify("an idea header", func() {
		c.Specify("can be parsed", func() {
			c.Specify("w/o an id", func() {
				headers := []string{
					"## [status] An Idea w/o an Id",
					"## [status] [] An Idea w/o an Id",
					"## [status] An Idea w/o an Id\n",
					"## [status] [] An Idea w/o an Id\n",
				}

				for _, header := range headers {
					status, id, name, err := parseHeader(header)
					c.Assume(err, IsNil)

					c.Expect(status, Equals, "status")
					c.Expect(id, Equals, uint(0))
					c.Expect(name, Equals, "An Idea w/o an Id")
				}
			})

			c.Specify("w/ an id", func() {
				headers := []string{
					"## [status] [1] An Idea w/ an Id",
					"## [status] [2] An Idea w/ an Id",
					"## [status] [3] An Idea w/ an Id\n",
				}

				for i, header := range headers {
					status, id, name, err := parseHeader(header)
					c.Assume(err, IsNil)

					c.Expect(status, Equals, "status")
					c.Expect(id, Equals, uint(i+1))
					c.Expect(name, Equals, "An Idea w/ an Id")
				}
			})
		})

		c.Specify("is invalid", func() {
			c.Specify("if the status isn't wrapped in []", func() {
				headers := []string{
					"## status [1] Header w/ an Id",
					"## status [] Header w/o an Id",
					"## status Header w/o an Id",
				}

				for _, header := range headers {
					_, _, _, err := parseHeader(header)
					c.Assume(err, Not(IsNil))
					c.Expect(err.Error(), Equals, "invalid idea header: status must be wrapped w/ []")
				}
			})
		})
	})

	c.Specify("an idea", func() {
		const someData = `
Some other markdowned text.
Doesn't matter what it is,
it will be skipped by the scanner.

## [active] An Idea w/o an Id
The newline before the next Idea should not be included

in the body of this Idea.

## [active] [] An Idea w/o an Id
The newline before the next Idea should not be included

in the body of this Idea.

## [active] [1] An Idea w/ an Id
The newline before the timestamp should not be included

in the body of this Idea.

Sun Jan 26 15:03:44 EST 2014
`
		c.Specify("can be scanned from some data", func() {
			iscan := NewIdeaScanner(strings.NewReader(someData))

			ideas := make([]*Idea, 0, 3)
			for iscan.Scan() {
				c.Assume(iscan.Err(), IsNil)
				ideas = append(ideas, iscan.Idea())
			}
			c.Assume(iscan.Err(), IsNil)
			c.Assume(len(ideas), Equals, 3)

			c.Specify("without an id", func() {
				for _, idea := range ideas[0:1] {
					c.Expect(*idea, Equals, Idea{
						IS_Active,
						0,
						"An Idea w/o an Id",
						`The newline before the next Idea should not be included

in the body of this Idea.
`,
					})
				}
			})

			c.Specify("with an id", func() {
				c.Expect(*ideas[2], Equals, Idea{
					IS_Active,
					1,
					"An Idea w/ an Id",
					`The newline before the timestamp should not be included

in the body of this Idea.
`,
				})
			})

			c.Specify("and will not include a timestamp as the final line", func() {
				c.Expect(ideas[len(ideas)-1].Body, Equals,
					`The newline before the timestamp should not be included

in the body of this Idea.
`)
			})
		})
	})
}
