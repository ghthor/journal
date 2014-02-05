package git

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
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

func DescribeCommit(c gospec.Context) {
	newChangesIn := func(d string) *Changes {
		d, err := ioutil.TempDir("_test/", d+"_")
		c.Assume(err, IsNil)
		c.Assume(Init(d), IsNil)
		return NewChangesIn(d)
	}

	makeSomeChangesIn := func(wd string, fileData []string) (changes []CommitableChange) {
		changes = make([]CommitableChange, 0, len(fileData))

		for _, data := range fileData {
			// Filename is a sha1 hash of the filedata
			h := sha1.New()
			_, err := io.Copy(h, strings.NewReader(data))
			c.Assume(err, IsNil)

			filename := hex.EncodeToString(h.Sum(nil))
			c.Assume(ioutil.WriteFile(filepath.Join(wd, filename), []byte(data), 0666), IsNil)

			changes = append(changes, ChangedFile(filename))
		}

		c.Assume(IsClean(wd), Not(IsNil))
		return
	}

	c.Specify("a collection of changes", func() {
		c.Specify("can be commited with a message", func() {
			changes := newChangesIn("changes_commit_test")

			for _, change := range makeSomeChangesIn(changes.WorkingDirectory(), []string{
				"file 1 data\n",
				"file 2 data\n",
			}) {
				changes.Add(change)
			}

			// Verify that the directory isn't clean
			o, err := Command(changes.WorkingDirectory(), "status", "-s").Output()
			c.Assume(err, IsNil)
			// And the changes haven't been `git add`ed
			c.Assume(string(o), Equals,
				`?? 0c6737cee25d5bb06f443e2e7daf229d78ad6b12
?? 8a63191cd06427fd6dfa4684080a5a5d40ae536c
`)

			changes.Msg = "Test Commit"
			c.Expect(changes.Commit(), IsNil)
			c.Expect(IsClean(changes.WorkingDirectory()), IsNil)

			o, err = Command(changes.WorkingDirectory(), "show", "--no-color", "--pretty=format:\"%s%b\"").Output()
			c.Assume(err, IsNil)

			c.Expect(string(o), Equals,
				`"Test Commit"
diff --git a/0c6737cee25d5bb06f443e2e7daf229d78ad6b12 b/0c6737cee25d5bb06f443e2e7daf229d78ad6b12
new file mode 100644
index 0000000..066428e
--- /dev/null
+++ b/0c6737cee25d5bb06f443e2e7daf229d78ad6b12
@@ -0,0 +1 @@
+file 2 data
diff --git a/8a63191cd06427fd6dfa4684080a5a5d40ae536c b/8a63191cd06427fd6dfa4684080a5a5d40ae536c
new file mode 100644
index 0000000..6a4a926
--- /dev/null
+++ b/8a63191cd06427fd6dfa4684080a5a5d40ae536c
@@ -0,0 +1 @@
+file 1 data
`)
		})
	})
}
