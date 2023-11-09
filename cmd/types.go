package cmd

import (
	"context"

	"github.com/jinzhu/copier"

	"github.com/jodydadescott/home-simplecert/client"
	"github.com/jodydadescott/home-simplecert/server"
	logger "github.com/jodydadescott/jody-go-logger"
)

type Logger = logger.Config
type ServerConfig = server.Config
type ClientConfig = client.Config

type Config struct {
	Logger *Logger       `json:"logger,omitempty" yaml:"logger,omitempty"`
	Server *ServerConfig `json:"server,omitempty" yaml:"server,omitempty"`
	Client *ClientConfig `json:"client,omitempty" yaml:"client,omitempty"`
}

// Clone return copy
func (t *Config) Clone() *Config {
	c := &Config{}
	copier.Copy(&c, &t)
	return c
}

type Runner interface {
	Run(ctx context.Context) error
}
