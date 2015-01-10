// DO NOT EDIT ** This file was generated with the bake tool ** DO NOT EDIT //

package case_0_static

var Files = map[string]string{
	"case_0/2014-01-01-0000-EST": `Wed Jan  1 00:00:00 EST 2014

# Commit Msg | Entry 1
Entry Body

entry_case_0
`,

	"case_0/2014-01-02-0000-EST": `Thu Jan  2 00:00:00 EST 2014

# Commit Msg | Entry 2
Entry Body

entry_case_1

Thu Jan  2 00:01:00 EST 2014
`,

	"case_0/2014-01-03-0000-EST": `Fri Jan  3 00:00:00 EST 2014

#~ Commit Msg
# Entry 3
Entry Body

entry_case_2

Fri Jan  3 00:01:00 EST 2014
`,

	"case_0/2014-01-04-0000-EST": `Sat Jan  4 00:00:00 EST 2014

#~ Commit Msg | Entry 4
Entry Body

entry_case_3

Sat Jan  4 00:01:00 EST 2014
`,

	"case_0/2014-01-05-0000-EST": `Sun Jan  5 00:00:00 EST 2014

#~ Commit Msg | Entry 5
Entry Body

entry_case_4

## [active] Active Idea
Idea Body

## [inactive] Inactive Idea
Idea Body

## [inactive] An Idea
Idea Body

Sun Jan  5 00:01:00 EST 2014
`,

	"case_0/2014-01-06-0000-EST": `Mon Jan  6 00:00:00 EST 2014

#~ Commit Msg | Entry 6
Entry Body

entry_case_4

## [inactive] Active Idea
Idea Body Updated

## [active] An Idea
Idea Body

Mon Jan  6 00:01:00 EST 2014
`,

	"case_0/2014-01-07-0000-EST": `Mon Jan  7 00:00:00 EST 2014

#~ Commit Msg | Entry 7
Entry Body

entry_case_4

## [active] An Idea
Idea Body

Mon Jan  7 00:01:00 EST 2014
`,

	"case_0.json": `{
    "directory":"case_0"
}
`,

	"case_0_fix_reflog/0": `journal - fix - begin
`,

	"case_0_fix_reflog/1": `journal - fix - moved all entries to entry/

diff --git a/2014-01-01-0000-EST b/2014-01-01-0000-EST
deleted file mode 100644
index d824b23..0000000
--- a/2014-01-01-0000-EST
+++ /dev/null
@@ -1,6 +0,0 @@
-Wed Jan  1 00:00:00 EST 2014
-
-# Commit Msg | Entry 1
-Entry Body
-
-entry_case_0
diff --git a/2014-01-02-0000-EST b/2014-01-02-0000-EST
deleted file mode 100644
index eaf21ec..0000000
--- a/2014-01-02-0000-EST
+++ /dev/null
@@ -1,8 +0,0 @@
-Thu Jan  2 00:00:00 EST 2014
-
-# Commit Msg | Entry 2
-Entry Body
-
-entry_case_1
-
-Thu Jan  2 00:01:00 EST 2014
diff --git a/2014-01-03-0000-EST b/2014-01-03-0000-EST
deleted file mode 100644
index 18aae5d..0000000
--- a/2014-01-03-0000-EST
+++ /dev/null
@@ -1,9 +0,0 @@
-Fri Jan  3 00:00:00 EST 2014
-
-#~ Commit Msg
-# Entry 3
-Entry Body
-
-entry_case_2
-
-Fri Jan  3 00:01:00 EST 2014
diff --git a/2014-01-04-0000-EST b/2014-01-04-0000-EST
deleted file mode 100644
index bc52ce1..0000000
--- a/2014-01-04-0000-EST
+++ /dev/null
@@ -1,8 +0,0 @@
-Sat Jan  4 00:00:00 EST 2014
-
-#~ Commit Msg | Entry 4
-Entry Body
-
-entry_case_3
-
-Sat Jan  4 00:01:00 EST 2014
diff --git a/2014-01-05-0000-EST b/2014-01-05-0000-EST
deleted file mode 100644
index 5906988..0000000
--- a/2014-01-05-0000-EST
+++ /dev/null
@@ -1,17 +0,0 @@
-Sun Jan  5 00:00:00 EST 2014
-
-#~ Commit Msg | Entry 5
-Entry Body
-
-entry_case_4
-
-## [active] Active Idea
-Idea Body
-
-## [inactive] Inactive Idea
-Idea Body
-
-## [inactive] An Idea
-Idea Body
-
-Sun Jan  5 00:01:00 EST 2014
diff --git a/2014-01-06-0000-EST b/2014-01-06-0000-EST
deleted file mode 100644
index ca4acf0..0000000
--- a/2014-01-06-0000-EST
+++ /dev/null
@@ -1,14 +0,0 @@
-Mon Jan  6 00:00:00 EST 2014
-
-#~ Commit Msg | Entry 6
-Entry Body
-
-entry_case_4
-
-## [inactive] Active Idea
-Idea Body Updated
-
-## [active] An Idea
-Idea Body
-
-Mon Jan  6 00:01:00 EST 2014
diff --git a/entry/2014-01-01-0000-EST b/entry/2014-01-01-0000-EST
new file mode 100644
index 0000000..d824b23
--- /dev/null
+++ b/entry/2014-01-01-0000-EST
@@ -0,0 +1,6 @@
+Wed Jan  1 00:00:00 EST 2014
+
+# Commit Msg | Entry 1
+Entry Body
+
+entry_case_0
diff --git a/entry/2014-01-02-0000-EST b/entry/2014-01-02-0000-EST
new file mode 100644
index 0000000..eaf21ec
--- /dev/null
+++ b/entry/2014-01-02-0000-EST
@@ -0,0 +1,8 @@
+Thu Jan  2 00:00:00 EST 2014
+
+# Commit Msg | Entry 2
+Entry Body
+
+entry_case_1
+
+Thu Jan  2 00:01:00 EST 2014
diff --git a/entry/2014-01-03-0000-EST b/entry/2014-01-03-0000-EST
new file mode 100644
index 0000000..18aae5d
--- /dev/null
+++ b/entry/2014-01-03-0000-EST
@@ -0,0 +1,9 @@
+Fri Jan  3 00:00:00 EST 2014
+
+#~ Commit Msg
+# Entry 3
+Entry Body
+
+entry_case_2
+
+Fri Jan  3 00:01:00 EST 2014
diff --git a/entry/2014-01-04-0000-EST b/entry/2014-01-04-0000-EST
new file mode 100644
index 0000000..bc52ce1
--- /dev/null
+++ b/entry/2014-01-04-0000-EST
@@ -0,0 +1,8 @@
+Sat Jan  4 00:00:00 EST 2014
+
+#~ Commit Msg | Entry 4
+Entry Body
+
+entry_case_3
+
+Sat Jan  4 00:01:00 EST 2014
diff --git a/entry/2014-01-05-0000-EST b/entry/2014-01-05-0000-EST
new file mode 100644
index 0000000..5906988
--- /dev/null
+++ b/entry/2014-01-05-0000-EST
@@ -0,0 +1,17 @@
+Sun Jan  5 00:00:00 EST 2014
+
+#~ Commit Msg | Entry 5
+Entry Body
+
+entry_case_4
+
+## [active] Active Idea
+Idea Body
+
+## [inactive] Inactive Idea
+Idea Body
+
+## [inactive] An Idea
+Idea Body
+
+Sun Jan  5 00:01:00 EST 2014
diff --git a/entry/2014-01-06-0000-EST b/entry/2014-01-06-0000-EST
new file mode 100644
index 0000000..ca4acf0
--- /dev/null
+++ b/entry/2014-01-06-0000-EST
@@ -0,0 +1,14 @@
+Mon Jan  6 00:00:00 EST 2014
+
+#~ Commit Msg | Entry 6
+Entry Body
+
+entry_case_4
+
+## [inactive] Active Idea
+Idea Body Updated
+
+## [active] An Idea
+Idea Body
+
+Mon Jan  6 00:01:00 EST 2014
`,

	"case_0_fix_reflog/10": `journal - fix - entry - format updated - entry/2014-01-04-0000-EST

diff --git a/entry/2014-01-04-0000-EST b/entry/2014-01-04-0000-EST
index bc52ce1..6b21e8d 100644
--- a/entry/2014-01-04-0000-EST
+++ b/entry/2014-01-04-0000-EST
@@ -1,6 +1,6 @@
 Sat Jan  4 00:00:00 EST 2014
 
-#~ Commit Msg | Entry 4
+# Commit Msg | Entry 4
 Entry Body
 
 entry_case_3
`,

	"case_0_fix_reflog/11": `journal - fix - entry - format updated - entry/2014-01-05-0000-EST

diff --git a/entry/2014-01-05-0000-EST b/entry/2014-01-05-0000-EST
index 5906988..625a6c9 100644
--- a/entry/2014-01-05-0000-EST
+++ b/entry/2014-01-05-0000-EST
@@ -1,17 +1,8 @@
 Sun Jan  5 00:00:00 EST 2014
 
-#~ Commit Msg | Entry 5
+# Commit Msg | Entry 5
 Entry Body
 
 entry_case_4
 
-## [active] Active Idea
-Idea Body
-
-## [inactive] Inactive Idea
-Idea Body
-
-## [inactive] An Idea
-Idea Body
-
 Sun Jan  5 00:01:00 EST 2014
`,

	"case_0_fix_reflog/12": `journal - fix - entry - format updated - entry/2014-01-06-0000-EST

diff --git a/entry/2014-01-06-0000-EST b/entry/2014-01-06-0000-EST
index ca4acf0..f55fe5f 100644
--- a/entry/2014-01-06-0000-EST
+++ b/entry/2014-01-06-0000-EST
@@ -1,14 +1,8 @@
 Mon Jan  6 00:00:00 EST 2014
 
-#~ Commit Msg | Entry 6
+# Commit Msg | Entry 6
 Entry Body
 
 entry_case_4
 
-## [inactive] Active Idea
-Idea Body Updated
-
-## [active] An Idea
-Idea Body
-
 Mon Jan  6 00:01:00 EST 2014
`,

	"case_0_fix_reflog/13": `journal - fix - completed
`,

	"case_0_fix_reflog/2": `journal - fix - idea directory store initialized

diff --git a/idea/active b/idea/active
new file mode 100644
index 0000000..e69de29
diff --git a/idea/nextid b/idea/nextid
new file mode 100644
index 0000000..d00491f
--- /dev/null
+++ b/idea/nextid
@@ -0,0 +1 @@
+1
`,

	"case_0_fix_reflog/3": `journal - fix - idea - created - 1 - src:entry/2014-01-05-0000-EST

diff --git a/idea/1 b/idea/1
new file mode 100644
index 0000000..a12bab8
--- /dev/null
+++ b/idea/1
@@ -0,0 +1,2 @@
+## [active] [1] Active Idea
+Idea Body
diff --git a/idea/active b/idea/active
index e69de29..d00491f 100644
--- a/idea/active
+++ b/idea/active
@@ -0,0 +1 @@
+1
diff --git a/idea/nextid b/idea/nextid
index d00491f..0cfbf08 100644
--- a/idea/nextid
+++ b/idea/nextid
@@ -1 +1 @@
-1
+2
`,

	"case_0_fix_reflog/4": `journal - fix - idea - created - 2 - src:entry/2014-01-05-0000-EST

diff --git a/idea/2 b/idea/2
new file mode 100644
index 0000000..be98c6b
--- /dev/null
+++ b/idea/2
@@ -0,0 +1,2 @@
+## [inactive] [2] Inactive Idea
+Idea Body
diff --git a/idea/nextid b/idea/nextid
index 0cfbf08..00750ed 100644
--- a/idea/nextid
+++ b/idea/nextid
@@ -1 +1 @@
-2
+3
`,

	"case_0_fix_reflog/5": `journal - fix - idea - created - 3 - src:entry/2014-01-05-0000-EST

diff --git a/idea/3 b/idea/3
new file mode 100644
index 0000000..8bc55b5
--- /dev/null
+++ b/idea/3
@@ -0,0 +1,2 @@
+## [inactive] [3] An Idea
+Idea Body
diff --git a/idea/nextid b/idea/nextid
index 00750ed..b8626c4 100644
--- a/idea/nextid
+++ b/idea/nextid
@@ -1 +1 @@
-3
+4
`,

	"case_0_fix_reflog/6": `journal - fix - idea - updated - 1 - src:entry/2014-01-06-0000-EST

diff --git a/idea/1 b/idea/1
index a12bab8..2a1f256 100644
--- a/idea/1
+++ b/idea/1
@@ -1,2 +1,2 @@
-## [active] [1] Active Idea
-Idea Body
+## [inactive] [1] Active Idea
+Idea Body Updated
diff --git a/idea/active b/idea/active
index d00491f..e69de29 100644
--- a/idea/active
+++ b/idea/active
@@ -1 +0,0 @@
-1
`,

	"case_0_fix_reflog/7": `journal - fix - idea - updated - 3 - src:entry/2014-01-06-0000-EST

diff --git a/idea/3 b/idea/3
index 8bc55b5..289b291 100644
--- a/idea/3
+++ b/idea/3
@@ -1,2 +1,2 @@
-## [inactive] [3] An Idea
+## [active] [3] An Idea
 Idea Body
diff --git a/idea/active b/idea/active
index e69de29..00750ed 100644
--- a/idea/active
+++ b/idea/active
@@ -0,0 +1 @@
+3
`,

	"case_0_fix_reflog/8": `journal - fix - entry - format updated - entry/2014-01-01-0000-EST

diff --git a/entry/2014-01-01-0000-EST b/entry/2014-01-01-0000-EST
index d824b23..90b0f8c 100644
--- a/entry/2014-01-01-0000-EST
+++ b/entry/2014-01-01-0000-EST
@@ -4,3 +4,5 @@ Wed Jan  1 00:00:00 EST 2014
 Entry Body
 
 entry_case_0
+
+Wed Jan  1 00:02:00 EST 2014
`,

	"case_0_fix_reflog/9": `journal - fix - entry - format updated - entry/2014-01-03-0000-EST

diff --git a/entry/2014-01-03-0000-EST b/entry/2014-01-03-0000-EST
index 18aae5d..7202c60 100644
--- a/entry/2014-01-03-0000-EST
+++ b/entry/2014-01-03-0000-EST
@@ -1,7 +1,6 @@
 Fri Jan  3 00:00:00 EST 2014
 
-#~ Commit Msg
-# Entry 3
+# Commit Msg | Entry 3
 Entry Body
 
 entry_case_2
`,

}
