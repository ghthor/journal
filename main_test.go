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

	c.Specify("the journal command", func() {
		c.Specify("will create a directory intialize an empty go repository", func() {
			c.Assume(os.RemoveAll("_test/journal"), IsNil)

			jc := exec.Command("./journal", "-edit=false", "-init", "-config=config.test.json")
			c.Expect(jc.Run(), IsNil)
			c.Expect("_test/journal", IsAGitRepository)
		})
	})
}
