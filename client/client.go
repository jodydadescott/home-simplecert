package client

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	logger "github.com/jodydadescott/jody-go-logger"
	"go.uber.org/zap"

	"github.com/jodydadescott/home-simplecert/libclient"
)

type Client struct {
	config *Config
	client *libclient.Client
}

func New(config *Config) (*Client, error) {

	if config == nil {
		panic("config is nil")
	}

	config = config.Clone()

	if config.Logger != nil {
		logger.SetConfig(config.Logger)
	}

	if config.Secret == "" {
		return nil, fmt.Errorf("secret is required")
	}

	if config.Server == "" {
		return nil, fmt.Errorf("server is required")
	}

	switch config.ModeType {

	case NormalModeType:

		if len(config.Domains) <= 0 {
			return nil, fmt.Errorf("one or more domains are required")
		}

		for _, domain := range config.Domains {

			if domain.Name == "" {
				return nil, fmt.Errorf("each domain must have a name")
			}

			if domain.CertFile == "" {
				return nil, fmt.Errorf("domain %s must have a certFile", domain.Name)
			}

			if domain.KeyFile == "" {
				return nil, fmt.Errorf("domain %s must have a keyFile", domain.Name)
			}

		}

	case SynologyModeType:
		if len(config.Domains) != 1 {
			return nil, fmt.Errorf("SynologyMode requires a single domain")
		}

	}

	return &Client{
		config: config,
		client: libclient.New(&libclient.Config{
			Secret:     config.Secret,
			Server:     config.Server,
			SkipVerify: config.SkipVerify,
		}),
	}, nil
}

func (t *Client) Run(ctx context.Context) error {

	fileExist := func(filename string) bool {

		_, err := os.Stat(filename)
		if err == nil {
			return true
		}

		return false
	}

	compare := func(file string, abytes []byte) bool {

		if !fileExist(file) {
			return false
		}

		bbytes, err := os.ReadFile(file)
		if err != nil {
			return false
		}

		return bytes.Equal(abytes, bbytes)
	}

	changed := func(domain *Domain, cert *CR) bool {

		result := false

		if compare(domain.CertFile, cert.GetCertPEM()) {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s CertFile unchanged", domain.Name))
			}
		} else {
			result = true
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s CertFile changed", domain.Name))
			}
		}

		if compare(domain.KeyFile, cert.GetKeyPEM()) {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s KeyFile unchanged", domain.Name))
			}
		} else {
			result = true
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s KeyFile changed", domain.Name))
			}
		}

		return result
	}

	process := func(domain *Domain, cert *CR) error {

		zap.L().Debug(fmt.Sprintf("Processing domain %s", domain.Name))

		err := os.MkdirAll(filepath.Dir(domain.CertFile), dirPerm)
		if err != nil {
			return err
		}

		err = os.WriteFile(domain.CertFile, cert.GetCertPEM(), filePerm)
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Dir(domain.KeyFile), dirPerm)
		if err != nil {
			return err
		}

		err = os.WriteFile(domain.KeyFile, cert.GetKeyPEM(), filePerm)
		if err != nil {
			return err
		}

		if domain.FullChain != "" {

			err := os.MkdirAll(filepath.Dir(domain.CertFile), dirPerm)
			if err != nil {
				return err
			}

			err = os.WriteFile(domain.FullChain, cert.GetCertPEM(), filePerm)
			if err != nil {
				return err
			}
		}

		if domain.Hook == nil {
			zap.L().Debug(fmt.Sprintf("Hook config not present for domain %s", domain.Name))
			return nil
		}

		zap.L().Debug(fmt.Sprintf("Running Hook for domain %s", domain.Name))
		cmd := exec.Command(domain.Hook.Name, domain.Hook.Args...)
		out, err := cmd.CombinedOutput()

		if werr, ok := err.(*exec.ExitError); ok {
			if s := werr.Error(); s != "0" {
				zap.L().Debug(fmt.Sprintf("Hook returned code %s error->%s, output->%s", s, string(out), string(out)))
				return err
			}
			zap.L().Debug(fmt.Sprintf("Hook returned zero code. Ouput->%s", string(out)))
		}

		return nil
	}

	run := func() error {

		zap.L().Debug("Processing domains")

		var errs *multierror.Error

		for _, domain := range t.config.Domains {

			cert, err := t.client.GetCert(domain.Name)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
				continue
			}

			if changed(domain, cert) {
				err = process(domain, cert)
				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
					continue
				}
			}

		}

		zap.L().Debug("Processing domains completed")

		return errs.ErrorOrNil()
	}

	runTick := func() {
		err := run()
		if err != nil {
			zap.L().Error(err.Error())
		}
	}

	runDaemon := func() {

		zap.L().Debug("Running as daemon")

		run()

		refreshInterval := t.config.RefreshInterval

		if refreshInterval > 0 {
			zap.L().Debug(fmt.Sprintf("Refresh Interval is %s (config)", refreshInterval.String()))
		} else {
			refreshInterval = defaultRefreshInterval
			zap.L().Debug(fmt.Sprintf("Refresh Interval is %s (default)", refreshInterval.String()))
		}

		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {

			select {

			case <-ctx.Done():
				zap.L().Debug("Shutting down on context")
				return

			case <-ticker.C:
				zap.L().Debug("Tick")
				return

			}

		}

	}

	runSynologyMode := func() error {

		domain := t.config.Domains[0]

		b, err := os.ReadFile(synologyDefaultFile)
		if err != nil {
			return err
		}

		cerFile := strings.ReplaceAll(string(b), "\n", "")
		certDir := filepath.Join(filepath.Dir(synologyDefaultFile), cerFile)

		hook := &Hook{
			Name: "systemctl",
		}
		hook.AddArgs("restart", "nginx.service")

		domain.CertFile = filepath.Join(certDir, synologyCertFile)
		domain.KeyFile = filepath.Join(certDir, synologyKeyFile)
		domain.FullChain = filepath.Join(certDir, synologyChainFile)

		cert, err := t.client.GetCert(domain.Name)
		if err != nil {
			return err
		}

		if changed(domain, cert) {
			err = process(domain, cert)
			if err != nil {
				return err
			}
		}

		return nil
	}

	defer t.client.Shutdown()

	switch t.config.ModeType {

	case NormalModeType:
		if t.config.Daemon {
			runTick()
			runDaemon()
			return nil
		}
		return run()

	case SynologyModeType:
		return runSynologyMode()

	}

	return nil
}
