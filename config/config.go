package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Directory string `json:"directory"`
}

func (c Config) ExpandEnv() Config {
	c.Directory = os.ExpandEnv(c.Directory)
	return c
}

// Will read the configuration from a .json file
// The directory will be expanded for Environment variables
func ReadFromFile(file string) (c Config, err error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return
	}

	c = c.ExpandEnv()
	return
}
