package idea

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git"
)

type IdeaIO struct {
	idea *Idea

	changes git.Commitable
	err     error
}

func SaveIn(d *DirectoryStore, iio *IdeaIO) error {
	iio.changes, iio.err = d.SaveIdea(iio.idea)
	return iio.err
}

func SaveNewIn(d *DirectoryStore, iio *IdeaIO) error {
	iio.changes, iio.err = d.SaveNewIdea(iio.idea)
	return iio.err
}

func UpdateIn(d *DirectoryStore, iio *IdeaIO) error {
	iio.changes, iio.err = d.UpdateIdea(*iio.idea)
	return iio.err
}

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
			ds, _, err := InitDirectoryStore(d)
			c.Assume(err, IsNil)
			c.Assume(ds, Not(IsNil))

			// Verify the directory has been initialized
			ds, err = NewDirectoryStore(d)
			c.Assume(err, IsNil)
			c.Assume(ds, Not(IsNil))

			return ds, d
		}

		activeIdeasIn := func(d *DirectoryStore) (activeIds []uint) {
			// Scan in the id's from the index file
			data, err := ioutil.ReadFile(filepath.Join(d.root, "active"))
			c.Assume(err, IsNil)

			scanner := bufio.NewScanner(bytes.NewReader(data))
			activeIds = make([]uint, 0, 3)

			for scanner.Scan() {
				var id uint
				n, err := fmt.Fscan(bytes.NewReader(scanner.Bytes()), &id)
				c.Assume(err, IsNil)
				c.Assume(n, Equals, 1)

				activeIds = append(activeIds, id)
			}
			return
		}

		c.Specify("can be initialized", func() {
			d := makeEmptyDirectory("directory_store_init")

			ds, commitable, err := InitDirectoryStore(d)
			c.Assume(err, IsNil)
			c.Expect(ds, Not(IsNil))

			c.Expect(ds.root, Equals, d)

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
				c.Expect(commitable.CommitMsg(), Equals, "idea directory store initialized")

				// Initialize and empty repo
				c.Assume(git.Init(d), IsNil)
				// Commit the directory store initialization
				c.Expect(git.Commit(commitable), IsNil)

				o, err := git.Command(d, "show", "--no-color", "--pretty=format:%s").Output()
				c.Assume(err, IsNil)
				c.Expect(string(o), Equals,
					`idea directory store initialized
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

		someIdeas := func() (ideas []*IdeaIO, activeIdeas []*IdeaIO, notActiveIdeas []*IdeaIO) {
			ideas = []*IdeaIO{{
				idea: &Idea{
					IS_Active,
					0,
					"A New Idea 1",
					"New Idea Body 1\nThis Idea is active\n",
				},
			}, {
				idea: &Idea{
					IS_Inactive,
					0,
					"A New Idea 2",
					"New Idea Body 2\nThis Idea is inactive\n",
				},
			}, {
				idea: &Idea{
					IS_Active,
					0,
					"A New Idea 3",
					"New Idea Body 3\nThis Idea is active\n",
				},
			}, {
				idea: &Idea{
					IS_Completed,
					0,
					"A Completed Idea",
					`This Idea has a really long body.
It is like this because I want to make sure the Update method works correctly
with extremely long bodies when the update is shorter.

The file should be truncated to reflect the shorter body.
`,
				},
			}}

			activeIdeas = make([]*IdeaIO, 0, 2)
			notActiveIdeas = make([]*IdeaIO, 0, 1)

			for _, iio := range ideas {
				// Build a parallel slice of active ideas
				if iio.idea.Status == IS_Active {
					activeIdeas = append(activeIdeas, iio)
				} else {
					notActiveIdeas = append(notActiveIdeas, iio)
				}
			}

			c.Assume(len(activeIdeas), Equals, 2)
			return
		}

		c.Specify("can create a new idea", func() {
			ds, d := makeDirectoryStore("directory_store_create")

			newIdeas, activeIdeas, notActiveIdeas := someIdeas()
			for _, iio := range newIdeas {
				c.Expect(SaveNewIn(ds, iio), IsNil)
				c.Expect(iio.changes, Not(IsNil))
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
					for _, iio := range newIdeas {
						c.Expect(iio.changes.Changes(), Contains, git.ChangedFile("nextid"))
					}
				})
			})

			c.Specify("by writing the idea to a file", func() {
				pathTo := func(idea *Idea) string {
					return filepath.Join(d, fmt.Sprintf("%d", idea.Id))
				}

				c.Specify("with the id as the filename", func() {

					for _, iio := range newIdeas {
						_, err := os.Stat(pathTo(iio.idea))
						c.Expect(!os.IsNotExist(err), IsTrue)
					}
				})

				for _, iio := range newIdeas {
					actualData, err := ioutil.ReadFile(pathTo(iio.idea))
					c.Assume(err, IsNil)

					r, err := NewIdeaReader(*iio.idea)
					c.Assume(err, IsNil)
					expectedData, err := ioutil.ReadAll(r)
					c.Assume(err, IsNil)

					c.Expect(string(actualData), Equals, string(expectedData))
				}

				c.Specify("and return a commitable change for the new idea file", func() {
					for _, iio := range newIdeas {
						c.Expect(iio.changes.Changes(), Contains, git.ChangedFile(fmt.Sprint(iio.idea.Id)))
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
					for _, iio := range activeIdeas {
						expectedActiveIds = append(expectedActiveIds, iio.idea.Id)
					}

					c.Expect(actualActiveIds, ContainsExactly, expectedActiveIds)

					c.Specify("and will return a commitable change for modifying the index", func() {
						for _, iio := range newIdeas {
							if iio.idea.Status == IS_Active {
								c.Expect(iio.changes.Changes(), Contains, git.ChangedFile("active"))
							} else {
								c.Expect(iio.changes.Changes(), Not(Contains), git.ChangedFile("active"))
							}
						}
					})
				})
			})

			c.Specify("and if the idea's status isn't active", func() {
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

					for _, iio := range notActiveIdeas {
						c.Expect(activeIds, Not(Contains), iio.idea.Id)
					}
				})
			})

			c.Specify("and returns a commitable change", func() {
				for _, iio := range newIdeas {
					c.Expect(iio.changes.CommitMsg(), Equals, fmt.Sprintf("idea - created - %d", iio.idea.Id))
				}
			})
		})

		c.Specify("can update an existing idea", func() {
			ds, d := makeDirectoryStore("directory_store_update")

			newIdeas, activeIdeas, notActiveIdeas := someIdeas()

			ideas := make([]*IdeaIO, 0, len(newIdeas))
			for _, iio := range newIdeas {
				// Fill the store w/ some ideas we can update
				c.Assume(SaveIn(ds, iio), IsNil)
				ideas = append(ideas, &IdeaIO{idea: iio.idea})
			}

			c.Specify("unless it hasn't been modified", func() {
				for _, iio := range ideas {
					c.Assume(UpdateIn(ds, iio), IsNil)
					c.Assume(iio.changes, IsNil)
				}
			})

			c.Specify("by writing the idea to the file", func() {
				for _, iio := range ideas {
					iio.idea.Body = fmt.Sprintf("Idea %d body has been modified\n", iio.idea.Id)
					c.Expect(UpdateIn(ds, iio), IsNil)
					c.Expect(iio.changes, Not(IsNil))
				}

				c.Specify("with the id as the filename", func() {
					for _, iio := range ideas {
						r, err := NewIdeaReader(*iio.idea)
						c.Assume(err, IsNil)

						expectedBytes, err := ioutil.ReadAll(r)
						c.Assume(err, IsNil)

						actualBytes, err := ioutil.ReadFile(filepath.Join(d, fmt.Sprint(iio.idea.Id)))
						c.Assume(err, IsNil)
						c.Expect(string(actualBytes), Equals, string(expectedBytes))
					}
				})

				c.Specify("and will return a commitable change for the modified idea file", func() {
					for _, iio := range ideas {
						c.Expect(iio.changes.Changes(), Contains, git.ChangedFile(fmt.Sprint(iio.idea.Id)))
						c.Expect(iio.changes.CommitMsg(), Equals, fmt.Sprintf("idea - updated - %d", iio.idea.Id))
					}
				})
			})

			c.Specify("and if the idea's status has changed", func() {
				c.Specify("to active", func() {
					nowActiveIdeas := make([]*IdeaIO, 0, len(notActiveIdeas))

					for _, iio := range notActiveIdeas {
						c.Assume(iio.idea.Status, Not(Equals), IS_Active)
						iio.idea.Status = IS_Active

						c.Expect(UpdateIn(ds, iio), IsNil)
						nowActiveIdeas = append(nowActiveIdeas, iio)
					}

					c.Specify("will add the idea's id to the active index", func() {
						activeIds := activeIdeasIn(ds)

						for _, iio := range nowActiveIdeas {
							c.Expect(activeIds, Contains, iio.idea.Id)
						}

						c.Specify("and will return a commitable change for modifying the index", func() {
							for _, iio := range nowActiveIdeas {
								c.Expect(iio.changes.Changes(), Contains, git.ChangedFile("active"))
								c.Expect(iio.changes.CommitMsg(), Equals, fmt.Sprintf("idea - updated - %d", iio.idea.Id))
							}
						})
					})
				})

				c.Specify("to not active", func() {
					nowNotActiveIdeas := make([]*IdeaIO, 0, len(activeIdeas))

					for _, iio := range activeIdeas {
						c.Assume(iio.idea.Status, Equals, IS_Active)
						iio.idea.Status = IS_Inactive

						c.Expect(UpdateIn(ds, iio), IsNil)
						nowNotActiveIdeas = append(nowNotActiveIdeas, iio)
					}

					c.Specify("will remove the id from the active index", func() {
						activeIds := activeIdeasIn(ds)

						for _, iio := range nowNotActiveIdeas {
							c.Expect(activeIds, Not(Contains), iio.idea.Id)
						}

						c.Specify("and will return a commitable change for modifying the index", func() {
							for _, iio := range nowNotActiveIdeas {
								c.Expect(iio.changes.Changes(), Contains, git.ChangedFile("active"))
								c.Expect(iio.changes.CommitMsg(), Equals, fmt.Sprintf("idea - updated - %d", iio.idea.Id))
							}
						})
					})
				})
			})

			c.Specify("and If the idea's status has NOT changed", func() {
				beforeUpdateIndex, err := ioutil.ReadFile(filepath.Join(d, "active"))
				c.Assume(err, IsNil)

				for _, iio := range ideas {
					iio.idea.Name = fmt.Sprintf("Idea %d Name", iio.idea.Id)
					iio.idea.Body = fmt.Sprintf("Idea %d Body\n", iio.idea.Id)
					c.Assume(UpdateIn(ds, iio), IsNil)
					c.Assume(iio.changes, Not(IsNil))
				}

				c.Specify("the active ids index will remain unchanged", func() {
					afterUpdateIndex, err := ioutil.ReadFile(filepath.Join(d, "active"))
					c.Assume(err, IsNil)
					c.Expect(string(afterUpdateIndex), Equals, string(beforeUpdateIndex))
				})

				c.Specify("will not return a commitable change for the active index", func() {
					for _, iio := range ideas {
						c.Expect(iio.changes.Changes(), Not(Contains), git.ChangedFile("active"))
						c.Expect(iio.changes.CommitMsg(), Equals, fmt.Sprintf("idea - updated - %d", iio.idea.Id))
					}
				})
			})
		})
	})
}
