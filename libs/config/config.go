package config

import (
	"log"
	"os"

	"github.com/ReCore-sys/bottombot2/libs/logging"
	"gopkg.in/yaml.v2"
)

// Configs is the config struct
type Configs struct {
	Token     string `yaml:"Token"`
	Prefix    string `yaml:"Prefix"`
	Ravenhost string `yaml:"ravenhost"`
	Ravenport int    `yaml:"ravenport"`
}

// Config Reads the config file and returns a Configs struct
func Config() Configs {
	// CFG is the global config struct.
	var CFG Configs
	// Read static/config.json and parse it into a config struct
	configFile, err := os.ReadFile("./static/config.yaml") // Read the config file
	if err != nil {

		println(1)
		log.Fatal(err)
	}
	err = yaml.Unmarshal(configFile, &CFG) // Parse the config file into the config struct
	if err != nil {

		println(4)
		logging.Log(err)
	}
	return CFG
}
