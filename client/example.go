package client

func ExampleConfig() *Config {

	refreshedHook := &Hook{
		Name: "systemctl",
	}

	refreshedHook.AddArgs("restart", "nginx")

	c := &Config{
		Notes:           ConfigNotes,
		Secret:          "the secret",
		Server:          "https://...",
		RefreshInterval: DefaultRefreshInterval,
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

	c.AddDomain(domain1, domain2)
	return c
}
