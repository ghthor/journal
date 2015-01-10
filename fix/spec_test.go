package fix

import (
	"testing"

	"github.com/ghthor/gospec"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeEntry)
	r.AddSpec(DescribeEntriesCollector)

	r.AddSpec(DescribeFixingCase0)

	r.AddSpec(DescribeAFixableJournal)

	gospec.MainGoTest(r, t)
}
