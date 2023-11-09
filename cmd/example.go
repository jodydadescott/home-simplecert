package cmd

import (
	logger "github.com/jodydadescott/jody-go-logger"

	"github.com/jodydadescott/home-simplecert/client"
	"github.com/jodydadescott/home-simplecert/server"
)

func ExampleClientConfig() *Config {
	return &Config{
		Client: client.ExampleConfig(),
		Logger: &Logger{
			LogLevel: logger.DebugLevel,
		},
	}
}

func ExampleServerConfig() *Config {
	return &Config{
		Server: server.ExampleConfig(),
		Logger: &Logger{
			LogLevel: logger.DebugLevel,
		},
	}
}
