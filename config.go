package main

import (
	"io"

	ipfs_config "github.com/ipfs/kubo/config"
)

type Config struct {
	cfg *ipfs_config.Config
}

func (c *Config) getConfig() (cfg *ipfs_config.Config) {
	return c.cfg
}

func NewDefaultConfig() (*Config, error) {
	cfg, err := initConfig(io.Discard, 2048)
	if err != nil {
		return nil, err
	}

	return &Config{cfg}, nil
}
