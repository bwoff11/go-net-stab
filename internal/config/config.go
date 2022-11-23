package config

import "github.com/spf13/viper"

type Configurcation struct {
	Interval  int
	Timeout   int
	Endpoints []string
}

var Config Configurcation

// Locations in which go-net-stab will look for a config.yml file
var configLocations = []string{
	"/etc/go-net-stab/",
	"$HOME/.config/go-net-stab/",
	"$HOME/.go-net-stab",
	".",
}

func LoadConfig() error {
	viper.SetConfigName("config")

	// Add possible locations for the config file
	for _, location := range configLocations {
		viper.AddConfigPath(location)
	}
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&Config); err != nil {
		return err
	}

	return nil
}
