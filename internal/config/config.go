package config

import (
	"encoding/json"
	"os"
)

type ServerConfig struct {
	Address       string `json:"address,omitempty"`
	StoreInterval string `json:"store_interval,omitempty"`
	StoreFile     string `json:"store_file,omitempty"`
	Restore       bool   `json:"restore,omitempty"`
	Key           string `json:"key,omitempty"`
	Dsn           string `json:"dsn,omitempty"`
	Crypt         string `json:"crypt,omitempty"`
	Subnet        string `json:"subnet,omitempty"`
}

const filename = "config.json"

func NewServerConfig() (*ServerConfig, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &ServerConfig{}

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type AgentConfig struct {
	Address        string `json:"address,omitempty"`
	PollInterval   string `json:"poll_interval,omitempty"`
	ReportInterval string `json:"report_interval,omitempty"`
	Key            string `json:"key,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	Crypto         string `json:"crypto,omitempty"`
}

func NewAgentConfig() (*AgentConfig, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &AgentConfig{}

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
