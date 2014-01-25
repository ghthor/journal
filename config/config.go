package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Directory string `json:"directory"`
}

func ReadFromFile(file string) (c Config, err error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &c)
	return
}
