package client

func ExampleConfig() *Config {

	refreshedHook := &Hook{
		Name: "systemctl",
	}

	refreshedHook.AddArgs("restart", "nginx")

	c := &Config{
		Notes:           "RefreshInterval is optional. It is only used if daemon is set to true.",
		Secret:          "the secret",
		Server:          "https://...",
		RefreshInterval: defaultRefreshInterval,
		Daemon:          true,
		SkipVerify:      false,
	}

	domain1 := &Domain{
		Name:     "example1.com",
		CertFile: "/path/to/certfile1.pem",
		KeyFile:  "/path/to/keyfile1.pem",
		Hook:     refreshedHook,
	}

	domain2 := &Domain{
		Name:     "example2.com",
		CertFile: "/path/to/certfile2.pem",
		KeyFile:  "/path/to/keyfile2.pem",
	}

	domain3 := &Domain{
		Name: "example3.com",
		Keystore: &Keystore{
			File:   "/path/to/keystore",
			Secret: "keystore_secret",
		},
	}

	c.AddDomain(domain1, domain2, domain3)
	return c
}

func ExampleSynologyConfig() *Config {

	domain := &Domain{
		Name: "example1.com",
	}

	c := &Config{
		Secret: "the secret",
		Server: "https://...",
	}

	c.AddDomain(domain)

	return c
}
