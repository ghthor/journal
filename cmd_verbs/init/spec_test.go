package init_test

import (
	"github.com/ghthor/gospec"
	"testing"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeInitCmd)

	gospec.MainGoTest(r, t)
}
