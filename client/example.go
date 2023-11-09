package client

func ExampleConfig() *Config {

	refreshedHook := &Hook{
		Name: "echo",
	}

	refreshedHook.AddArgs("touch")
	refreshedHook.AddArgs("/tmp/refreshed")

	c := &Config{
		Notes:           "RefreshInterval is optional. It is only used if daemon=true. Setting a PreHook with the FailOnErr will cause the domain fetch to stop. If running in daemon mode it will remain running. Setting FailOnErr on a PostHook will have not effect",
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

	c.AddDomain(domain1, domain2)
	return c
}

func ExampleSynologyConfig() *Config {

	domain := &Domain{
		Name: "example1.com",
	}

	c := &Config{
		Secret:     "the secret",
		Server:     "https://...",
		SkipVerify: false,
	}

	c.AddDomain(domain)

	return c
}
