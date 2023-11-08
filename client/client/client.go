package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-multierror"
	logger "github.com/jodydadescott/jody-go-logger"
	"go.uber.org/zap"

	"github.com/jodydadescott/home-simplecert/client/libclient"
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

	run := func() error {

		zap.L().Debug("Processing domains")

		var errs *multierror.Error

		if t.config.PreHook != nil && t.config.PreHook.Enabled {

			zap.L().Debug("Executing PreHook")

			cmd := exec.Command(t.config.PreHook.Name, t.config.PreHook.Args...)
			err := cmd.Run()

			if err != nil {
				if t.config.PreHook.FailOnErr {
					return err
				}
				errs = multierror.Append(errs, fmt.Errorf("PreHook %w", err))
			}

		} else {
			if logger.Trace {
				zap.L().Debug("Prehook is not enabled")
			}
		}

		for _, domain := range t.config.Domains {

			zap.L().Debug(fmt.Sprintf("Processing domain %s", domain.Name))

			cert, err := t.client.GetCert(domain.Name)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
				continue
			}

			err = os.MkdirAll(filepath.Dir(domain.CertFile), dirPerm)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
				continue
			}

			err = os.MkdirAll(filepath.Dir(domain.KeyFile), dirPerm)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
				continue
			}

			err = os.WriteFile(domain.CertFile, cert.GetCertPEM(), filePerm)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
			}

			err = os.WriteFile(domain.KeyFile, cert.GetKeyPEM(), filePerm)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
			}

		}

		if t.config.PostHook != nil && t.config.PostHook.Enabled {

			zap.L().Debug("Executing PostHook")

			cmd := exec.Command(t.config.PostHook.Name, t.config.PostHook.Args...)
			err := cmd.Run()

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("PostHook %w", err))
			}

		} else {
			if logger.Trace {
				zap.L().Debug("PostHook is not enabled")
			}
		}

		zap.L().Debug("Processing domains completed")

		return errs.ErrorOrNil()
	}

	runLog := func() {
		err := run()
		if err != nil {
			zap.L().Error(err.Error())
		}
	}

	if !t.config.Daemon {
		err := run()
		t.client.Shutdown()
		return err
	}

	zap.L().Debug("Running as daemon")

	runLog()

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
			return nil

		case <-ticker.C:
			zap.L().Debug("Tick")
			runLog()

		}

	}

}
