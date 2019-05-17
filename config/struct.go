package config

// Config represents whole configuration file structure.
type Config struct {
	NATS Nats `yaml:"nats"`
}

// Nats represents NATS connection configuration.
type Nats struct {
	ConnectionString string `yaml:"connection_string"`
}
