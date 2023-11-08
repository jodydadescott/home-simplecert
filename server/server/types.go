package server

import (
	"github.com/jinzhu/copier"
	"github.com/jodydadescott/home-simplecert/types"
	logger "github.com/jodydadescott/jody-go-logger"
)

type AuthRequest = types.AuthRequest
type TokenResponse = types.TokenResponse
type CertResponse = types.CertResponse
type CR = types.CR
type SimpleMessage = types.SimpleMessage
type Logger = logger.Config

type Config struct {
	Notes         string    `json:"notes,omitempty" yaml:"notes,omitempty"`
	Logger        *Logger   `json:"logger,omitempty" yaml:"logger,omitempty"`
	PrimaryDomain *Domain   `json:"primaryDomain,omitempty" yaml:"primaryDomain,omitempty"`
	Domains       []*Domain `json:"domains,omitempty" yaml:"domains,omitempty"`
	Email         string    `json:"email,omitempty" yaml:"email,omitempty"`
	CacheDir      string    `json:"cacheDir,omitempty" yaml:"cacheDir,omitempty"`
	Secret        string    `json:"secret,omitempty" yaml:"secret,omitempty"`
}

// Clone return copy
func (t *Config) Clone() *Config {
	c := &Config{}
	copier.Copy(&c, &t)
	return c
}

func (t *Config) AddDomain(domains ...*Domain) *Config {
	t.Domains = append(t.Domains, domains...)
	return t
}

type Domain struct {
	Name    string   `json:"name,omitempty" yaml:"name,omitempty"`
	Aliases []string `json:"aliases,omitempty" yaml:"aliases,omitempty"`
}

func (t *Domain) AddAliases(aliases ...string) *Domain {
	t.Aliases = append(t.Aliases, aliases...)
	return t
}

// Clone return copy
func (t *Domain) Clone() *Domain {
	c := &Domain{}
	copier.Copy(&c, &t)
	return c
}

type HTTPDebug struct {
	Request  string `json:"request,omitempty" yaml:"request,omitempty"`
	Response string `json:"response,omitempty" yaml:"response,omitempty"`
}
