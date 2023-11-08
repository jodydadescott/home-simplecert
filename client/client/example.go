package client

import (
	"time"

	logger "github.com/jodydadescott/jody-go-logger"
)

func ExampleConfig() *Config {

	preHook := &Hook{
		Enabled: true,
		Name:    "echo",
	}

	preHook.AddArgs("hello")

	postHook := &Hook{
		Enabled: true,
		Name:    "echo",
	}

	postHook.AddArgs("goodbye")

	c := &Config{
		Notes: `
		RefreshInterval is optional. It is only used if daemon=true. Setting a PreHook
		with the FailOnErr will cause the domain fetch to stop. If running in daemon mode
		it will remain running. Setting FailOnErr on a PostHook will have not effect.
		`,

		Secret:          "the secret",
		Server:          "https://...",
		RefreshInterval: time.Hour * 400,
		Daemon:          true,
		SkipVerify:      false,
		PreHook:         preHook,
		PostHook:        postHook,
		Logger: &Logger{
			LogLevel: logger.DebugLevel,
		},
	}

	domain1 := &Domain{
		Name:     "example1.com",
		CertFile: "/path/to/certfile1.pem",
		KeyFile:  "/path/to/keyfile1.pem",
	}

	domain2 := &Domain{
		Name:     "example2.com",
		CertFile: "/path/to/certfile2.pem",
		KeyFile:  "/path/to/keyfile2.pem",
	}

	c.Secret = "thesecret"
	c.Server = "https://..."
	c.SkipVerify = false
	c.AddDomain(domain1, domain2)
	return c
}
