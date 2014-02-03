package git

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	. "github.com/ghthor/journal/git/gittest"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestIntegrationSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeGitIntegration)

	gospec.MainGoTest(r, t)
}

func DescribeGitIntegration(c gospec.Context) {
	c.Specify("a git repository will be created", func() {
		d, err := ioutil.TempDir("", "git_integration_test")
		c.Assume(err, IsNil)

		defer func(dir string) {
			c.Expect(os.RemoveAll(dir), IsNil)
		}(d)

		d = path.Join(d, "a_git_repo")
		c.Expect(GitInit(d), IsNil)
		c.Expect(d, IsAGitRepository)

		c.Specify("and will be clean", func() {
			c.Expect(GitIsClean(d), IsNil)
		})

		testFile := path.Join(d, "test_file")
		c.Assume(ioutil.WriteFile(testFile, []byte("some data\n"), 0666), IsNil)

		c.Specify("and will be dirty", func() {
			c.Expect(GitIsClean(d).Error(), Equals, "directory is dirty")
		})

		c.Specify("and will add a file", func() {
			c.Expect(GitAdd(d, testFile), IsNil)
			o, err := GitCommand(d, "status", "-s").Output()
			c.Assume(err, IsNil)
			c.Expect(string(o), Equals, "A  test_file\n")
		})

		c.Specify("and will commit all staged changes", func() {
			c.Assume(GitAdd(d, testFile), IsNil)
			c.Expect(GitCommitAll(d, "a commit msg"), IsNil)

			o, err := GitCommand(d, "show", "--no-color", "--pretty=format:\"%s%b\"").Output()
			c.Assume(err, IsNil)

			c.Expect(string(o), Equals, `"a commit msg"
diff --git a/test_file b/test_file
new file mode 100644
index 0000000..4268632
--- /dev/null
+++ b/test_file
@@ -0,0 +1 @@
+some data
`)
		})
	})
}
