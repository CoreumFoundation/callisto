package bridge

import (
	"gopkg.in/yaml.v3"
)

// Config contains the configuration about the bridge module
type Config struct {
	ContractAddress string `yaml:"contract_address"`
}

// NewConfig returns a new Config instance
func NewConfig(contractAddress string) *Config {
	return &Config{
		ContractAddress: contractAddress,
	}
}

func ParseConfig(bz []byte) (*Config, error) {
	type T struct {
		Config *Config `yaml:"bridge"`
	}
	var cfg T
	err := yaml.Unmarshal(bz, &cfg)
	return cfg.Config, err
}
