package client

import (
	"time"

	"github.com/jinzhu/copier"
	logger "github.com/jodydadescott/jody-go-logger"

	"github.com/jodydadescott/home-simplecert/types"
)

type CR = types.CR
type Logger = logger.Config

type Config struct {
	Notes           string        `json:"notes,omitempty" yaml:"notes,omitempty"`
	Secret          string        `json:"secret" yaml:"secret"`
	Server          string        `json:"server" yaml:"server"`
	SkipVerify      bool          `json:"skipVerify" yaml:"skipVerify"`
	Domains         []*Domain     `json:"domains,omitempty" yaml:"domains,omitempty"`
	RefreshInterval time.Duration `json:"refreshInterval,omitempty" yaml:"refreshInterval,omitempty"`
	Daemon          bool          `json:"daemon,omitempty" yaml:"daemon,omitempty"`
	IgnoreOSType    bool          `json:"ignoreOSType" yaml:"ignoreOSType"`
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
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
	DomainName string `json:"domainName,omitempty" yaml:"domainName,omitempty"`
	KeyFile    string `json:"keyFile,omitempty" yaml:"keyFile,omitempty"`
	CertFile   string `json:"certFile,omitempty" yaml:"certFile,omitempty"`
	FullChain  string `json:"fullChain,omitempty" yaml:"fullChain,omitempty"`
	Hook       *Hook  `json:"hook,omitempty" yaml:"hook,omitempty"`
}

// Clone return copy
func (t *Domain) Clone() *Domain {
	c := &Domain{}
	copier.Copy(&c, &t)
	return c
}

type Hook struct {
	Name string   `json:"name" yaml:"name"`
	Args []string `json:"args,omitempty" yaml:"args,omitempty"`
}

// Clone return copy
func (t *Hook) Clone() *Hook {
	c := &Hook{}
	copier.Copy(&c, &t)
	return c
}

func (t *Hook) GetCmd() string {
	cmd := t.Name
	for _, arg := range t.Args {
		cmd = cmd + " " + arg
	}
	return cmd
}

// Clone return copy
func (t *Hook) AddArgs(args ...string) *Hook {
	t.Args = append(t.Args, args...)
	return t
}
