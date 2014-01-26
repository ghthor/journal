package main

import (
	"errors"
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func tmpGitRepository(prefix string) (dir string, err error) {
	dir, err = ioutil.TempDir("", prefix)
	if err != nil {
		return "", err
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", err
	}

	gitInitCmd := exec.Command(gitPath, "init", dir)
	err = gitInitCmd.Run()
	if err != nil {
		return "", err
	}

	return
}

func IsAGitRepository(dir interface{}, _ interface{}) (match bool, pos gospec.Message, neg gospec.Message, err error) {
	d, ok := dir.(string)
	if !ok {
		return false, pos, neg, errors.New("directory is not a string")
	}

	// Check if jd exists and is a Directory
	if info, err := os.Stat(d); !os.IsNotExist(err) {
		if !info.IsDir() {
			return false, pos, neg, errors.New(fmt.Sprintf("%s is not a directory", d))
		}
	} else {
		// jd doesn't exist
		return false, pos, neg, err
	}

	pos = gospec.Messagef(fmt.Sprintf("%s directory doesn't exist", path.Join(d, ".git/")), "%s is a git repository", d)
	neg = gospec.Messagef(fmt.Sprintf("%s directory does exist", path.Join(d, ".git/")), "%s is NOT a git repository", d)

	// Check if a .git directory exists
	if info, err := os.Stat(path.Join(d, ".git/")); !os.IsNotExist(err) {
		if !info.IsDir() {
			return false, pos, neg, nil
		}
	} else {
		// .git directory doesn't exist
		return false, pos, neg, nil
	}

	pos = gospec.Messagef(d, "%s is a git repository", d)
	neg = gospec.Messagef(d, "%s is NOT a git repository", d)

	match = true
	return
}

func DescribeGitIntegration(c gospec.Context) {
	c.Specify("a temporary git repository can be created", func() {
		d, err := tmpGitRepository("git_integration_test")
		c.Expect(err, IsNil)
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

		c.Expect(os.RemoveAll(d), IsNil)
	})
}
