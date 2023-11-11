package cmd

import (
	logger "github.com/jodydadescott/jody-go-logger"
)

// func ExampleClientConfigNormal() *Config {
// 	return &Config{
// 		Client: client.ExampleConfig(),
// 		Logger: &Logger{
// 			LogLevel: logger.DebugLevel,
// 		},
// 	}
// }

// func ExampleClientConfigNormal() *Config {
// 	return &Config{
// 		Client: client.ExampleConfig(),
// 		Logger: &Logger{
// 			LogLevel: logger.DebugLevel,
// 		},
// 	}
// }

func exampleConfig() *Config {
	return &Config{
		Logger: &Logger{
			LogLevel: logger.DebugLevel,
		},
	}
}
