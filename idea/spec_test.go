package idea

import (
	"github.com/ghthor/gospec"
	"testing"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeIdea)
	r.AddSpec(DescribeIdeaStore)

	gospec.MainGoTest(r, t)
}
