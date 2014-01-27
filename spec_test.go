package main

import (
	"github.com/ghthor/gospec"
	"testing"
)

func TestUnitSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeNewCmd)
	r.AddSpec(DescribeWatchCmd)
	r.AddSpec(DescribeIdea)

	gospec.MainGoTest(r, t)
}

func TestIntegrationSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeGitIntegration)

	gospec.MainGoTest(r, t)
}

func TestExecutableSpec(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeJournalCommand)

	gospec.MainGoTest(r, t)
}
