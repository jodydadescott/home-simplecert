package libclient

import (
	"github.com/jinzhu/copier"
	hashauthserver "github.com/jodydadescott/simple-go-hash-auth/server"

	"github.com/jodydadescott/home-simplecert/types"
)

type AuthRequest = hashauthserver.AuthRequest
type TokenResponse = types.TokenResponse
type Token = hashauthserver.Token
type CertResponse = types.CertResponse
type CR = types.CR

type Config struct {
	Secret     string `json:"secret" yaml:"secret"`
	Server     string `json:"server" yaml:"server"`
	SkipVerify bool   `json:"skipVerify" yaml:"skipVerify"`
}

// Clone return copy
func (t *Config) Clone() *Config {
	c := &Config{}
	copier.Copy(&c, &t)
	return c
}
