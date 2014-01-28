package config

import (
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
	"os"
	"testing"
)

func TestSpecs(t *testing.T) {
	r := gospec.NewRunner()

	r.AddSpec(DescribeConfigLoading)

	gospec.MainGoTest(r, t)
}

func DescribeConfigLoading(c gospec.Context) {
	c.Specify("a config", func() {
		c.Specify("can be stored in a json file", func() {
			expectedConfig := Config{
				"a/path/to/a/git/directory",
			}

			config, err := ReadFromFile("config.example.json")
			c.Assume(err, IsNil)
			c.Expect(config, Equals, expectedConfig)
		})

		c.Specify("stores a directory", func() {
			c.Specify("that will have it's environment varibles expanded", func() {
				expectedConfig := Config{
					"expanded_path_to/_test",
				}
				c.Assume(os.Setenv("A_ENV_VAR", "expanded_path_to"), IsNil)

				config, err := ReadFromFile("config.env_test.json")
				c.Assume(err, IsNil)
				c.Expect(config, Equals, expectedConfig)
			})
		})
	})
}
