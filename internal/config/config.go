package config

import (
	"errors"
	"io/ioutil"
	"net"
	"time"

	"gopkg.in/yaml.v2"
)

// Configuration structure
type Configuration struct {
	Interval  time.Duration `yaml:"interval"`
	Timeout   time.Duration `yaml:"timeout"`
	Port      string        `yaml:"port"`
	Localhost string        `yaml:"localhost"`
	Endpoints []Endpoint    `yaml:"endpoints"`
}

// Endpoint structure
type Endpoint struct {
	Hostname string        `yaml:"hostname"`
	Address  string        `yaml:"address"`
	Location string        `yaml:"location"`
	Interval time.Duration `yaml:"interval"`
}

func New() *Configuration {
	return &Configuration{}
}

func (p *Configuration) Load() error {
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configFile, p)
	if err != nil {
		return err
	}

	return nil
}

func (p *Configuration) Validate() error {
	if p.Interval <= 0 {
		return errors.New("interval must be a positive duration")
	}

	if p.Timeout <= 0 {
		return errors.New("timeout must be a positive duration")
	}

	if p.Port == "" {
		return errors.New("port must not be empty")
	}

	if p.Localhost == "" {
		return errors.New("localhost must not be empty")
	}

	if !isValidIP(p.Localhost) {
		return errors.New("localhost must be a valid IP address")
	}

	if len(p.Endpoints) == 0 {
		return errors.New("at least one endpoint must be specified")
	}

	for _, endpoint := range p.Endpoints {
		if err := validateEndpoint(endpoint); err != nil {
			return err
		}
	}

	return nil
}

func validateEndpoint(endpoint Endpoint) error {
	if endpoint.Hostname == "" {
		return errors.New("Endpoint hostname must not be empty")
	}

	if endpoint.Address == "" {
		return errors.New("Endpoint address must not be empty")
	}

	if !isValidIP(endpoint.Address) {
		return errors.New("Endpoint address must be a valid IP address")
	}

	if endpoint.Location == "" {
		return errors.New("Endpoint location must not be empty")
	}

	if endpoint.Interval <= 0 {
		return errors.New("Endpoint interval must be a positive duration")
	}

	return nil
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
