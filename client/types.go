package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/jodydadescott/home-simplecert/types"
	logger "github.com/jodydadescott/jody-go-logger"
)

type CR = types.CR
type Logger = logger.Config

func (t *Config) ParseModeType(s string) error {

	switch strings.ToLower(s) {

	case "":
		t.ModeType = NormalModeType

	case string(NormalModeType):
		t.ModeType = NormalModeType

	case string(SynologyModeType):
		t.ModeType = SynologyModeType

	default:
		return fmt.Errorf("ModeType %s is invalid", s)

	}

	return nil
}

type Config struct {
	Notes           string        `json:"notes,omitempty" yaml:"notes,omitempty"`
	Secret          string        `json:"secret" yaml:"secret"`
	Server          string        `json:"server" yaml:"server"`
	SkipVerify      bool          `json:"skipVerify" yaml:"skipVerify"`
	Domains         []*Domain     `json:"domains,omitempty" yaml:"domains,omitempty"`
	RefreshInterval time.Duration `json:"refreshInterval,omitempty" yaml:"refreshInterval,omitempty"`
	Daemon          bool          `json:"daemon,omitempty" yaml:"daemon,omitempty"`
	Logger          *Logger       `json:"logger,omitempty" yaml:"logger,omitempty"`
	ModeType        ModeType      `json:"modeType,omitempty" yaml:"modeType,omitempty"`
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
	Name      string `json:"name" yaml:"name"`
	KeyFile   string `json:"keyFile" yaml:"keyFile"`
	CertFile  string `json:"certFile" yaml:"certFile"`
	FullChain string `json:"fullChain" yaml:"fullChain"`
	Hook      *Hook  `json:"hook,omitempty" yaml:"hook,omitempty"`
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

// Clone return copy
func (t *Hook) AddArgs(args ...string) *Hook {
	t.Args = append(t.Args, args...)
	return t
}
