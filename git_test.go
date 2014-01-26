	return dir, GitInit(dir)
	c.Specify("a git repository will be created", func() {
		d, err := ioutil.TempDir("", "git_integration_test")
		c.Assume(err, IsNil)

		defer func(dir string) {
			c.Expect(os.RemoveAll(dir), IsNil)
		}(d)

		d = path.Join(d, "a_git_repo")
		c.Expect(GitInit(d), IsNil)
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