package bridge

import (
	"gopkg.in/yaml.v3"
)

// Config contains the configuration about the bridge module
type Config struct {
	XrplContractAddress string `yaml:"xrpl_contract_address"`
}

// NewConfig returns a new Config instance
func NewConfig(xrplContractAddress string) *Config {
	return &Config{
		XrplContractAddress: xrplContractAddress,
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
