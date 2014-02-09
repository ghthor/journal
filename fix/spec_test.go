package fix

import (
	"github.com/ghthor/gospec"
	"testing"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeEntry)

	r.AddSpec(DescribeJournalCase0)

	gospec.MainGoTest(r, t)
}
