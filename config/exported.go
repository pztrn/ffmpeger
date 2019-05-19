package config

import (
	// stdlib
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"strings"

	// other
	"gopkg.in/yaml.v2"
)

var (
	configPathRaw string
	Cfg           *Config
)

// Initialize initializes package.
func Initialize() {
	log.Println("Initializing configuration...")

	flag.StringVar(&configPathRaw, "conf", "", "Path to configuration file.")

	Cfg = &Config{}
}

// Load loads configuration into memory and parses it into Config struct.
func Load() error {
	if configPathRaw == "" {
		return errors.New("No configuration file path defined! See '-h'!")
	}

	log.Println("Loading configuration from file:", configPathRaw)

	// Replace home directory if "~" was specified.
	if strings.Contains(configPathRaw, "~") {
		u, err := user.Current()
		if err != nil {
			// Well, I don't know how to test this.
			return errors.New("Failed to get current user's data: " + err.Error())
		}

		configPathRaw = strings.Replace(configPathRaw, "~", u.HomeDir, 1)
	}

	// Get absolute path to configuration file.
	configPath, err := filepath.Abs(configPathRaw)
	if err != nil {
		// Can't think of situation when it's testable.
		return errors.New("Failed to get real configuration file path:" + err.Error())
	}

	// Read it.
	configFileData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.New("Failed to load configuration file data:" + err.Error())
	}

	// Parse it.
	err1 := yaml.Unmarshal(configFileData, Cfg)
	if err1 != nil {
		return errors.New("Failed to parse configuration file:" + err1.Error())
	}

	log.Printf("Configuration file parsed: %+v\n", Cfg)
	return nil
}
