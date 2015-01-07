package fix

import (
	"github.com/ghthor/gospec"
	"testing"
)

func TestUnitSpecs(t *testing.T) {

	r := gospec.NewRunner()

	r.AddSpec(DescribeEntry)
	r.AddSpec(DescribeEntriesCollector)

	cleanupFn, err := initCase0()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupFn()
	r.AddSpec(DescribeJournalCase0)

	gospec.MainGoTest(r, t)
}
