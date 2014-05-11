package main

import (
	"fmt"
	"github.com/gonuts/yaml"
	"io/ioutil"
	"os"
)

type ProcessConfig struct {
	Name    string   `name`
	Command string   `command`
	Streams []string `streams`
	Restart bool     `restart`
	After   []string `after`
}

type Config struct {
	Processes []ProcessConfig `processes`
}

func ParseConfig(filename string) (config *Config, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, NewError("Could not open file: "+filename, err)
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, NewError("Could not read file: "+filename, err)
	}
	config = new(Config)
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, NewError("Could not parse YAML config: "+filename, err)
	}
	return
}

func validateConfig(config *Config) error {
	for _, process := range config.Processes {
		for _, depend := range process.After {
			n, t := parseAfter(depend)
			if !(t == "started" || t == "finished") {
				return NewError(fmt.Sprintf("Invalid dependency type %s for %s", t, process.Name), nil)
			}
			valid := false
			for _, process := range config.Processes {
				if n == process.Name {
					valid = true
					break
				}
			}
			if !valid {
				return NewError(fmt.Sprintf("Non-existent dependency %s for %s", n, process.Name), nil)
			}
		}
	}
	return nil
}
