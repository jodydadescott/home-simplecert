package client

import (
	"time"

	"github.com/jinzhu/copier"
	logger "github.com/jodydadescott/jody-go-logger"
	hashauthserver "github.com/jodydadescott/simple-go-hash-auth/server"

	"github.com/jodydadescott/home-simplecert/types"
)

type Config struct {
	Notes           string        `json:"notes" yaml:"notes"`
	Secret          string        `json:"secret" yaml:"secret"`
	Server          string        `json:"server" yaml:"server"`
	SkipVerify      bool          `json:"skipVerify" yaml:"skipVerify"`
	Domains         []*Domain     `json:"domains" yaml:"domains"`
	RefreshInterval time.Duration `json:"refreshInterval" yaml:"refreshInterval"`
	Daemon          bool          `json:"daemon" yaml:"daemon"`
	Logger          *Logger       `json:"logger,omitempty" yaml:"logger,omitempty"`
	PreHook         *Hook         `json:"preHook,omitempty" yaml:"preHook,omitempty"`
	PostHook        *Hook         `json:"postHook,omitempty" yaml:"postHook,omitempty"`
}

// Clone return copy
func (t *Config) Clone() *Config {
	c := &Config{}
	copier.Copy(&c, &t)
	return c
}

// Clone return copy
func (t *Config) AddDomain(domains ...*Domain) *Config {
	t.Domains = append(t.Domains, domains...)
	return t
}

type Domain struct {
	Name     string `json:"name" yaml:"name"`
	KeyFile  string `json:"keyFile" yaml:"keyFile"`
	CertFile string `json:"certFile" yaml:"certFile"`
}

// Clone return copy
func (t *Domain) Clone() *Domain {
	c := &Domain{}
	copier.Copy(&c, &t)
	return c
}

type Hook struct {
	Name      string   `json:"name" yaml:"name"`
	Enabled   bool     `json:"enabled" yaml:"enabled"`
	FailOnErr bool     `json:"failOnErr" yaml:"failOnErr"`
	Args      []string `json:"args,omitempty" yaml:"args,omitempty"`
}

// Clone return copy
func (t *Hook) Clone() *Hook {
	c := &Hook{}
	copier.Copy(&c, &t)
	return c
}

// Clone return copy
func (t *Hook) AddArgs(args ...string) *Hook {
	t.Args = append(t.Args, args...)
	return t
}

type Logger = logger.Config
type AuthRequest = hashauthserver.AuthRequest
type TokenResponse = types.TokenResponse
type Token = hashauthserver.Token
type CertResponse = types.CertResponse
type CR = types.CR
