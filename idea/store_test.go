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
	"os"
	"path/filepath"
)

func DescribeIdeaStore(c gospec.Context) {
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

			c.Expect(id.root, Equals, d)

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

			c.Specify("and returns a commitable change", func() {
				for _, ni := range newIdeas {
					c.Expect(ni.changes.CommitMsg(), Equals, fmt.Sprintf("IDEA - %d - Created", ni.idea.Id))
				}
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
