package main

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"github.com/ghthor/journal/git/gittest"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var BuildExecutableOnce sync.Once
var MkTestDirectory sync.Once

func DescribeJournalCommand(c gospec.Context) {
	BuildExecutableOnce.Do(func() {
		goCmdPath, err := exec.LookPath("go")
		if err != nil {
			log.Fatal(err)
		}

		err = exec.Command(goCmdPath, "build").Run()
		if err != nil {
			log.Fatal(err)
		}
	})

	MkTestDirectory.Do(func() {
		if err := os.Mkdir("_test", 0755); err != nil {
			if !os.IsExist(err) {
				// Only fatal if is isn't that the directory already exists
				log.Fatal(err)
			}
		}
	})

	c.Specify("the journal command", func() {
		c.Specify("will use environment variable expansion for filepaths", func() {
			wd, err := os.Getwd()
			c.Assume(err, IsNil)

			wd, err = filepath.Abs(wd)
			c.Assume(err, IsNil)
			c.Assume(os.Setenv("JOURNAL_PKG_DIR", wd), IsNil)

			c.Specify("for the configuration path", func() {
				td := "_test/config.env_exp"
				c.Assume(ioutil.WriteFile("_test/config.env_exp.json",
					[]byte(`{"directory":"`+td+`"}`), 0666), IsNil)

				jc := exec.Command("./journal", "-edit=false", "-init",
					"-config=$JOURNAL_PKG_DIR/_test/config.env_exp.json", "new")
				c.Expect(jc.Run(), IsNil)
			})

			c.Specify("for the directory stored in the configuration", func() {
				td := "$JOURNAL_PKG_DIR/_test/config.directory.env_exp"
				c.Assume(ioutil.WriteFile("_test/config.directory.env_exp.json",
					[]byte(`{"directory":"`+td+`"}`), 0666), IsNil)

				jc := exec.Command("./journal", "-edit=false", "-init",
					"-config=_test/config.directory.env_exp.json", "new")
				c.Expect(jc.Run(), IsNil)
			})
		})

		c.Specify("will create a directory and intialize an empty go repository", func() {
			td := "_test/journal-init"
			c.Assume(ioutil.WriteFile("_test/config.journal-init.json", []byte(`{"directory":"`+td+`"}`), 0666), IsNil)

			jc := exec.Command("./journal", "-edit=false", "-init", "-config=_test/config.journal-init.json", "new")
			c.Expect(jc.Run(), IsNil)
			c.Expect(td, gittest.IsAGitRepository)
		})
	})
}
