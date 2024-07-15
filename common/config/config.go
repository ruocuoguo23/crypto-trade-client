package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type TLSConfig struct {
	TLSDisable    bool
	TLSCaFile     string
	TLSCertFile   string
	TLSKeyFile    string
	TLSMinVersion string
	ServerName    string
}

// Config represents the structure of the configuration file
type Config struct {
	Chains []Chain `yaml:"Chains"`
}

// Chain represents a single chain configuration
type Chain struct {
	Name       string `yaml:"name"`
	URL        string `yaml:"url"`
	PrivateKey string `yaml:"privateKey"`
}

func LoadConfig(configPath string) (map[string]Chain, error) {
	// Read the YAML configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Unmarshal the YAML data into the Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config data: %v", err)
	}

	// Create a map to store the chains with the chain name as the key
	chainMap := make(map[string]Chain)
	for _, chain := range config.Chains {
		chainMap[chain.Name] = chain
	}

	return chainMap, nil
}
