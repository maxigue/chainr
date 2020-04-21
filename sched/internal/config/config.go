// Package config contains the configuration of the scheduler.
// It exposes a function to read the configuration into the configuration
// structure.
package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Port int
}

var dfltConfig Configuration = Configuration{
	Port: 8080,
}

// Load returns the configuration structure, constructed from the
// configuration file at the given location.
// The configuration can only contain a subset of the configuration. Any
// configuration entry that is not set in the configuration file will be
// replaced by its default value.
// If the configuration loading fails (the file does not exist or is not
// readable), the default configuration is returned, along with an error
// describing the problem.
func Load(filename string) (Configuration, error) {
	log.Println("Reading configuration from", filename)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return dfltConfig, err
	}

	c := dfltConfig
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return dfltConfig, err
	}

	return c, nil
}
