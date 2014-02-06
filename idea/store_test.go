package idea

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
	"io/ioutil"
	"os"
	"path/filepath"
)

func DescribeIdeaStore(c gospec.Context) {
	c.Specify("a directory store", func() {
		makeEmptyDirectory := func(prefix string) string {
			d, err := ioutil.TempDir("_test", prefix+"_")
			c.Assume(err, IsNil)
			return d
		}

		makeDirectoryStore := func(prefix string) (*DirectoryStore, string) {
			d := makeEmptyDirectory(prefix)

			// Verify the directory isn't an DirectoryStore
			_, err := NewDirectoryStore(d)
			c.Assume(IsInvalidDirectoryStoreError(err), IsTrue)

			// Initialize the directory
			id, _, err := InitDirectoryStore(d)
			c.Assume(err, IsNil)
			c.Assume(id, Not(IsNil))

			// Verify the directory has been initialized
			id, err = NewDirectoryStore(d)
			c.Assume(err, IsNil)
			c.Assume(id, Not(IsNil))

			return id, d
		}

		c.Specify("can be initialized", func() {
			d := makeEmptyDirectory("directory_store_init")

			id, commitable, err := InitDirectoryStore(d)
			c.Assume(err, IsNil)
			c.Expect(id, Not(IsNil))

			c.Expect(id.root, Equals, d)

			c.Specify("only once", func() {
				_, _, err = InitDirectoryStore(d)
				c.Expect(err, Equals, ErrInitOnExistingDirectoryStore)
			})

			c.Specify("and the modifications made during initialization are commitable", func() {
				c.Expect(commitable, Not(IsNil))
				c.Expect(commitable.WorkingDirectory(), Equals, d)
				c.Expect(commitable.Changes(), ContainsAll, []git.ChangedFile{
					git.ChangedFile("nextid"),
					git.ChangedFile("active"),
				})
				c.Expect(commitable.CommitMsg(), Equals, "directory store initialized")

				// Initialize and empty repo
				c.Assume(git.Init(d), IsNil)
				// Commit the directory store initialization
				c.Expect(git.Commit(commitable), IsNil)

				o, err := git.Command(d, "show", "--no-color", "--pretty=format:\"%s%b\"").Output()
				c.Assume(err, IsNil)
				c.Expect(string(o), Equals,
					`"directory store initialized"
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

		c.Specify("that is empty", func() {
			_, d := makeDirectoryStore("directory_store_spec")

			c.Specify("has a next available id of 1", func() {
				data, err := ioutil.ReadFile(filepath.Join(d, "nextid"))
				c.Expect(err, IsNil)
				c.Expect(string(data), Equals, "1\n")
			})

			c.Specify("contains an empty index of active ideas", func() {
				fs, err := os.Stat(filepath.Join(d, "active"))
				c.Expect(err, IsNil)
				c.Expect(fs.Size(), Equals, int64(0))
			})

			c.Specify("contains no ideas", func() {
				fileCount := 0
				filepath.Walk(d, func(path string, fi os.FileInfo, err error) error {
					if path != d {
						fileCount++
					}
					return nil
				})
				c.Expect(fileCount, Equals, 2)
			})
		})

		c.Specify("contains ideas stored in a files", func() {
			c.Specify("with the id as the filename", func() {
			})
		})

		c.Specify("can create a new idea", func() {
			id, d := makeDirectoryStore("directory_store_create")

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

			activeIdeas := make([]*newIdea, 0, 2)

			for _, ni := range newIdeas {
				// Build a parallel slice of active ideas
				if ni.idea.Status == IS_Active {
					activeIdeas = append(activeIdeas, ni)
				}

				changes, err := id.SaveNewIdea(ni.idea)
				c.Assume(err, IsNil)
				c.Assume(changes, Not(IsNil))
				ni.changes = changes
			}

			c.Assume(len(activeIdeas), Equals, 2)

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
				c.Specify("will add the idea's id to the active index", func() {
					data, err := ioutil.ReadFile(filepath.Join(d, "active"))
					c.Assume(err, IsNil)

					actualActiveIds := make([]uint, 0, len(activeIdeas))
					scanner := bufio.NewScanner(bytes.NewReader(data))

					for scanner.Scan() {
						var id uint
						_, err := fmt.Fscan(bytes.NewReader(scanner.Bytes()), &id)
						c.Assume(err, IsNil)

						actualActiveIds = append(actualActiveIds, id)
					}

					expectedActiveIds := make([]uint, 0, len(activeIdeas))
					for _, ni := range activeIdeas {
						expectedActiveIds = append(expectedActiveIds, ni.idea.Id)
					}

					c.Expect(actualActiveIds, ContainsExactly, expectedActiveIds)

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
