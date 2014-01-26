package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"io/ioutil"
	"os"
	"path"
	"text/template"
	"time"
)

func DescribeNewCmd(c gospec.Context) {
	jd, err := tmpGitRepository("journal_test")
	c.Assume(err, IsNil)

	defer func() {
		err := os.RemoveAll(jd)
		c.Assume(err, IsNil)
	}()

	c.Specify("the `new` command", func() {
		const fileData = `
#~ git commit msg
# Subject/Title
Some data that shouldn't be modified
`

		tmpl, err := template.New(jd).Parse(fileData)
		c.Assume(err, IsNil)

		c.Specify("will append the current time after editting is completed", func() {
			j, err := newEntry(jd, tmpl, nil, nil, &Command{})
			c.Assume(err, IsNil)

			actualData, err := ioutil.ReadFile(path.Join(jd, j.Filename))
			c.Assume(err, IsNil)

			c.Expect(string(actualData), Equals, fmt.Sprintf("%s\n%s\n", fileData, j.ClosedAt))
		})

		c.Specify("will commit the entry to the git repository", func() {
			j, err := newEntry(jd, entryTmpl, func() time.Time {
				return time.Time{}
			}, nil, &Command{})
			c.Assume(err, IsNil)

			o, err := GitCommand(jd, "show", "--no-color", "--pretty=format:\"%s%b\"").Output()
			c.Assume(err, IsNil)

			actualData := bytes.NewBuffer(o)
			expectedData := bytes.NewBuffer(make([]byte, 0, 1024))

			commitMsgTmpl, err := template.New("commitMsgTmpl").Parse(
				`"Event(will be used as commit message)"
diff --git a/{{.Filename}} b/{{.Filename}}
new file mode 100644
index 0000000..951ffa6
--- /dev/null
+++ b/{{.Filename}}
@@ -0,0 +1,7 @@
+{{.OpenedAt}}
+
+#~ Event(will be used as commit message)
+# Subject
+TODO Make this some random quote or something stupid
+
+{{.ClosedAt}}
`)
			c.Assume(err, IsNil)
			c.Assume(commitMsgTmpl.Execute(expectedData, j), IsNil)

			c.Expect(actualData.String(), Equals, expectedData.String())

			// Helps with debugging the test
			// Outputs the first line that doesn't match
			actualDataSc, expectedDataSc := bufio.NewScanner(actualData), bufio.NewScanner(expectedData)
			for actualDataSc.Scan() && expectedDataSc.Scan() {
				c.Expect(actualDataSc.Text(), Equals, expectedDataSc.Text())
				if actualDataSc.Text() != expectedDataSc.Text() {
					break
				}
			}
		})

		c.Specify("will fail", func() {
			c.Specify("if the journal directory has a dirty git repository", func() {
				c.Assume(ioutil.WriteFile(path.Join(jd, "dirty"), []byte("some data"), os.FileMode(0600)), IsNil)
				_, err := newEntry(jd, entryTmpl, nil, nil, &Command{})
				c.Expect(err, Not(IsNil))
			})
		})
	})
}
