package idea

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	c.Specify("an idea reader", func() {
		c.Specify("can read an idea", func() {
			ideas := []Idea{{
				"status",
				0,
				"An Idea w/o an Id",
				"An Idea body of text\n",
			}, {
				"status",
				2,
				"An Idea w/ an Id",
				"An Idea body of text\n",
			}}

			c.Specify("without an id", func() {
				idea := ideas[0]
				ideaReader, err := NewIdeaReader(idea)
				c.Assume(err, IsNil)

				dst := bytes.NewBuffer(make([]byte, 0, 1024))

				expected := `## [status] An Idea w/o an Id
An Idea body of text
`
				n, err := io.Copy(dst, ideaReader)
				c.Expect(err, IsNil)
				c.Expect(int(n), Equals, len(expected))
				c.Expect(dst.String(), Equals, expected)
			})

			c.Specify("with an id", func() {
				idea := ideas[1]
				ideaReader, err := NewIdeaReader(idea)
				c.Assume(err, IsNil)

				dst := bytes.NewBuffer(make([]byte, 0, 1024))

				n, err := io.Copy(dst, ideaReader)
				c.Expect(err, IsNil)

				expected := "## [status] [2] An Idea w/ an Id\nAn Idea body of text\n"
				c.Expect(int(n), Equals, len(expected))
				c.Expect(dst.String(), Equals, expected)
			})
		})
	})

	c.Specify("an idea directory", func() {
		makeEmptyDirectory := func(prefix string) string {
			d, err := ioutil.TempDir("_test", prefix+"_")
			c.Assume(err, IsNil)
			return d
		}

		makeIdeaDirectory := func(prefix string) (*IdeaDirectory, string) {
			d := makeEmptyDirectory(prefix)

			// Verify the directory isn't an IdeaDirectory
			_, err := NewIdeaDirectory(d)
			c.Assume(IsInvalidIdeaDirectoryError(err), IsTrue)

			// Initialize the directory
			id, _, err := InitIdeaDirectory(d)
			c.Assume(err, IsNil)
			c.Assume(id, Not(IsNil))

			// Verify the directory has been initialized
			id, err = NewIdeaDirectory(d)
			c.Assume(err, IsNil)
			c.Assume(id, Not(IsNil))

			return id, d
		}

		c.Specify("can be initialized", func() {
			d := makeEmptyDirectory("idea_directory_init")

			id, commitable, err := InitIdeaDirectory(d)
			c.Assume(err, IsNil)
			c.Expect(id, Not(IsNil))

			c.Expect(id.directory, Equals, d)

			c.Specify("only once", func() {
				_, _, err = InitIdeaDirectory(d)
				c.Expect(err, Equals, ErrInitOnExistingIdeaDirectory)
			})

			c.Specify("and the modifications made during initialization are commitable", func() {
				c.Expect(commitable, Not(IsNil))
				c.Expect(commitable.WorkingDirectory(), Equals, d)
				c.Expect(commitable.Changes(), ContainsAll, []git.ChangedFile{
					git.ChangedFile("nextid"),
					git.ChangedFile("active"),
				})
				c.Expect(commitable.CommitMsg(), Equals, "idea directory initialized")

				// Initialize and empty repo
				c.Assume(git.Init(d), IsNil)
				// Commit the idea directory initialization
				c.Expect(git.Commit(commitable), IsNil)

				o, err := git.Command(d, "show", "--no-color", "--pretty=format:\"%s%b\"").Output()
				c.Assume(err, IsNil)
				c.Expect(string(o), Equals,
					`"idea directory initialized"
diff --git a/active b/active
new file mode 100644
index 0000000..e69de29
diff --git a/nextid b/nextid
new file mode 100644
index 0000000..d00491f
--- /dev/null
+++ b/nextid
@@ -0,0 +1 @@
+1
`)
			})
		})

		c.Specify("contains an index of the next available id", func() {
			_, d := makeIdeaDirectory("idea_directory_spec")

			data, err := ioutil.ReadFile(filepath.Join(d, "nextid"))
			c.Expect(err, IsNil)
			c.Expect(string(data), Equals, "1\n")
		})

		c.Specify("contains an index of active ideas", func() {
			_, d := makeIdeaDirectory("idea_directory_spec")

			_, err := os.Stat(filepath.Join(d, "active"))
			c.Expect(err, IsNil)
		})

		c.Specify("contains ideas stored in a files", func() {
			c.Specify("with the id as the filename", func() {
			})
		})

		c.Specify("can create a new idea", func() {
			id, d := makeIdeaDirectory("idea_directory_create")

			type newIdea struct {
				changes git.Commitable
				idea    *Idea
			}

			newIdeas := []*newIdea{{
				idea: &Idea{
					IS_Active,
					0,
					"A New Idea",
					"Some body text\n",
				},
			}, {
				idea: &Idea{
					IS_Inactive,
					0,
					"Another New Idea",
					"That isn't active\n",
				},
			}, {
				idea: &Idea{
					IS_Active,
					0,
					"Another Active New Idea",
					"That should be active\n",
				},
			}}

			for _, ni := range newIdeas {
				changes, err := id.SaveNewIdea(ni.idea)
				c.Assume(err, IsNil)
				c.Assume(changes, Not(IsNil))
				ni.changes = changes
			}

			c.Specify("by assigning the next available id to the idea", func() {
				c.Expect(newIdeas[0].idea.Id, Equals, uint(1))
				c.Expect(newIdeas[1].idea.Id, Equals, uint(2))
			})

			c.Specify("by incrementing the next available id", func() {
				data, err := ioutil.ReadFile(filepath.Join(d, "nextid"))
				c.Assume(err, IsNil)

				var nextId uint
				n, err := fmt.Fscan(bytes.NewReader(data), &nextId)
				c.Expect(err, IsNil)
				c.Expect(n, Equals, 1)
				c.Expect(nextId, Equals, uint(len(newIdeas)+1))

				c.Specify("and will return a commitable change for modifying the next available id", func() {
					for _, ni := range newIdeas {
						c.Expect(ni.changes.Changes(), Contains, git.ChangedFile("nextid"))
					}
				})
			})

			c.Specify("by writing the idea to a file", func() {
				pathTo := func(idea *Idea) string {
					return filepath.Join(d, fmt.Sprintf("%d", idea.Id))
				}

				c.Specify("with the id as the filename", func() {

					for _, ni := range newIdeas {
						_, err := os.Stat(pathTo(ni.idea))
						c.Expect(!os.IsNotExist(err), IsTrue)
					}
				})

				for _, ni := range newIdeas {
					actualData, err := ioutil.ReadFile(pathTo(ni.idea))
					c.Assume(err, IsNil)

					r, err := NewIdeaReader(*ni.idea)
					c.Assume(err, IsNil)
					expectedData, err := ioutil.ReadAll(r)
					c.Assume(err, IsNil)

					c.Expect(string(actualData), Equals, string(expectedData))
				}

				c.Specify("and return a commitable change for the new idea file", func() {
					for _, ni := range newIdeas {
						c.Expect(ni.changes.Changes(), Contains, git.ChangedFile(fmt.Sprint(ni.idea.Id)))
					}
				})
			})

			c.Specify("and if the idea's status is active", func() {
				activeIdeas := make([]*Idea, 0, 2)
				for _, ni := range newIdeas {
					if ni.idea.Status == IS_Active {
						activeIdeas = append(activeIdeas, ni.idea)
					}
				}
				c.Assume(len(activeIdeas), Equals, 2)

				c.Specify("will add the idea's id to the active index", func() {
					data, err := ioutil.ReadFile(filepath.Join(d, "active"))
					c.Assume(err, IsNil)

					r := bytes.NewReader(data)

					var id uint
					activeIdeaIds := make([]uint, 0, len(activeIdeas))

					// Can just use fmt.Fscan because we know how many lines there are
					for i := 0; i < len(activeIdeas); i++ {
						_, err := fmt.Fscan(r, &id)
						c.Assume(err, IsNil)

						activeIdeaIds = append(activeIdeaIds, id)
					}

					_, err = fmt.Fscan(r, &id)
					c.Assume(err, Equals, io.EOF)

					c.Specify("and will return a commitable change for modifying the index", func() {
						for _, ni := range newIdeas {
							if ni.idea.Status == IS_Active {
								c.Expect(ni.changes.Changes(), Contains, git.ChangedFile("active"))
							} else {
								c.Expect(ni.changes.Changes(), Not(Contains), git.ChangedFile("active"))
							}
						}
					})
				})
			})

			c.Specify("and if the idea's status isn't active", func() {
				notActiveIdeas := make([]*Idea, 0, 1)
				for _, ni := range newIdeas {
					if ni.idea.Status != IS_Active {
						notActiveIdeas = append(notActiveIdeas, ni.idea)
					}
				}
				c.Assume(len(notActiveIdeas), Equals, 1)

				c.Specify("will not add the idea's id to the active index", func() {
					// Collect the id's from the index file
					data, err := ioutil.ReadFile(filepath.Join(d, "active"))
					c.Assume(err, IsNil)

					// Using a scanner because I don't know how many there are
					scanner := bufio.NewScanner(bytes.NewReader(data))
					activeIds := make([]uint, 0, len(newIdeas))

					for scanner.Scan() {
						var id uint
						_, err := fmt.Fscan(bytes.NewReader(scanner.Bytes()), &id)
						c.Assume(err, IsNil)
						activeIds = append(activeIds, id)
					}
					c.Assume(len(activeIds), Equals, 2)

					for _, idea := range notActiveIdeas {
						c.Expect(activeIds, Not(Contains), idea.Id)
					}
				})
			})
		})

		c.Specify("can update an existing idea", func() {
			c.Specify("by writing the idea to the file", func() {
				c.Specify("with the id as the filename", func() {
				})

				c.Specify("and will return a commitable change for the modified idea file", func() {
				})
			})

			c.Specify("and if the idea's status is active", func() {
				c.Specify("will add the idea's id to the active index", func() {
					c.Specify("and will return a commitable change for modifying the index", func() {
					})
				})
			})

			c.Specify("and if the idea's status isn't active", func() {
				c.Specify("will not add the idea's id to the active index", func() {
				})
			})
		})
	})
}
