package main

import (
	"os"

	"github.com/goccy/go-yaml"

	"github.com/lukasdietrich/ical-proxy/internal/proxy"
)

type httpConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Config struct {
	HTTP      httpConfig            `yaml:"http"`
	Calendars proxy.CalendarMuxSpec `yaml:"calendars"`
}

func defaultConfig() Config {
	return Config{
		HTTP: httpConfig{
			Host: "0.0.0.0",
			Port: "8080",
		},
	}
}

func parseConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	defer f.Close()

	config := defaultConfig()
	return &config, yaml.NewDecoder(f, yaml.Strict()).Decode(&config)
}
