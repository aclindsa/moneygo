package main

import (
	"fmt"
	"gopkg.in/gcfg.v1"
)

type Config struct {
	MoneyGo struct {
		Fcgi           bool   // whether to serve FCGI (HTTP by default if false)
		Base_directory string // base directory for serving files out of
		Port           int    // port to serve API/files on
	}
}

func readConfig(filename string) (*Config, error) {
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %s", err)
	}
	return &cfg, nil
}
