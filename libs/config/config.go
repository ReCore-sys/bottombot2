package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Configs is the config struct
type Configs struct {
	Token     string
	Prefix    string
	Ravenhost string
	Ravenport int
}

// Config Reads the config file and returns a Configs struct
func Config() Configs {
	// CFG is the global config struct.
	var CFG Configs
	// Read static/config.json and parse it into a config struct
	configFile, err := ioutil.ReadFile("./static/config.json") // Read the config file
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(configFile, &CFG) // Parse the config file into the config struct
	if err != nil {
		log.Println(err)
	}
	return CFG
}
