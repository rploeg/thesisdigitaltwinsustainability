package main

import (
	"github.com/iot-for-all/iiot-oee/pkg/models"
	"github.com/iot-for-all/iiot-oee/pkg/simulating"
)

type (
	Config struct {
		LogLevel string `json:"logLevel"` // logging level for the application
		LogsDir  string `json:"logsDir"`  // directory into which logs are written
	}

	config struct {
		Logger      Config                    `json:"logger"`
		Application models.CentralApplication `json:"application"`
		Plant       []simulating.Plant        `json:"plant"`
	}
)

func newConfig() *config {
	return &config{
		Logger: Config{
			LogLevel: "Debug",
			LogsDir:  "./logs",
		},
		Application: models.CentralApplication{},
		Plant:       []simulating.Plant{},
	}
}
