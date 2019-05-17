package config

import (
	// stdlib
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

	flag.StringVar(&configPathRaw, "conf", "", "Path to configuration file, should be absolute.")
}

// Load loads configuration into memory and parses it into Config struct.
func Load() {
	if configPathRaw == "" {
		log.Fatalln("No configuration file path defined! See '-h'!")
	}

	log.Println("Loading configuration from file:", configPathRaw)

	// Replace home directory if "~" was specified.
	if strings.Contains(configPathRaw, "~") {
		u, err := user.Current()
		if err != nil {
			log.Fatalln("Failed to get current user's data:", err.Error())
		}

		configPathRaw = strings.Replace(configPathRaw, "~", u.HomeDir, 1)
	}

	// Get absolute path to configuration file.
	configPath, err := filepath.Abs(configPathRaw)
	if err != nil {
		log.Fatalln("Failed to get real configuration file path:", err.Error())
	}

	// Read it.
	configFileData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalln("Failed to load configuration file data:", err.Error())
	}

	// Parse it.
	Cfg = &Config{}
	err1 := yaml.Unmarshal(configFileData, Cfg)
	if err1 != nil {
		log.Fatalln("Failed to parse configuration file:", err1.Error())
	}

	log.Printf("Configuration file parsed: %+v\n", Cfg)
}
