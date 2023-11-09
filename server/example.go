package server

func ExampleConfig() *Config {

	c := &Config{
		Notes:    "If the global secret is not set then each domain secret must be set. If the global secret is set and a domain secret is set the domain secret overrides the domain secret",
		Email:    "nobody@example.com",
		CacheDir: "letsencrypt",
		Secret:   "secret",
	}

	c.PrimaryDomain = &Domain{
		Name: "example.com",
	}

	c.PrimaryDomain.AddAliases("www.example.com")

	domain1 := &Domain{
		Name: "example1.com",
	}

	domain1.AddAliases("www.example1.com", "api.example1.com")

	domain2 := &Domain{
		Name: "example2.com",
	}

	domain2.AddAliases("www.example2.com")

	c.AddDomain(domain1)
	c.AddDomain(domain2)

	return c
}
