package entry

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/idea"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func init() {
	var err error
	if _, err = os.Stat("_test/"); os.IsNotExist(err) {
		err = os.Mkdir("_test/", 0755)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func DescribeAnEntry(c gospec.Context) {
	td, err := ioutil.TempDir("_test", "entry_")
	c.Assume(err, IsNil)

	ne := New(td)

	c.Specify("an entry", func() {
		c.Specify("can be opened", func() {
			c.Specify("at a specific time", func() {
				t := time.Date(2006, time.January, 1, 1, 0, 0, 0, time.UTC)
				oe, err := ne.Open(func() time.Time {
					return t
				}, nil)
				c.Assume(err, IsNil)
				c.Expect(oe.OpenedAt(), Equals, t)
			})

			c.Specify("with a list of ideas", func() {
				ideas := []idea.Idea{{
					Name:   "Active Idea",
					Status: idea.IS_Active,
					Body:   "Some text\n.",
				}, {
					Name:   "Another Idea",
					Status: idea.IS_Active,
					Body:   "Some other text\n",
				}}
				oe, err := ne.Open(time.Now, ideas)

				c.Assume(err, IsNil)
				for i, idea := range oe.Ideas() {
					c.Expect(idea, Equals, ideas[i])
				}
			})
		})

		c.Specify("that is open", func() {
			c.Specify("is a file", func() {
			})
			c.Specify("will have the time opened as the first line of the entry", func() {
			})
			c.Specify("will have a list of ideas appended to the entry", func() {
			})
			c.Specify("can be editted by a text editor", func() {
			})
			c.Specify("can be closed", func() {
			})
		})

		c.Specify("that is closed", func() {
			c.Specify("will have all ideas removed from the entry", func() {
			})
			c.Specify("will have the time closed as the last line of the entry", func() {
			})
			c.Specify("can be commited to the git repository", func() {
			})
		})
	})
}
