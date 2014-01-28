package main

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"log"
	"os"
	"os/exec"
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
		c.Specify("will create a directory intialize an empty go repository", func() {
			td := "_test/journal-init"
			c.Assume(ioutil.WriteFile("_test/config.journal-init.json", []byte(`{"directory":"`+td+`"}`), 0666), IsNil)

			jc := exec.Command("./journal", "-edit=false", "-init", "-config=_test/config.journal-init.json", "new")
			c.Expect(jc.Run(), IsNil)
			c.Expect(td, IsAGitRepository)
		})
	})
}
