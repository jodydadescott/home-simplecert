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
	osType OSType
}

func New(config *Config) (*Client, error) {

	if config == nil {
		panic("config is nil")
	}

	config = config.Clone()

	if config.Secret == "" {
		return nil, fmt.Errorf("secret is required")
	}

	if config.Server == "" {
		return nil, fmt.Errorf("server is required")
	}

	processDomain := func(domain *Domain) error {

		if domain.Name == "" {
			return fmt.Errorf("each domain must have a name")
		}

		if domain.CertFile == "" {
			return fmt.Errorf("domain %s must have a certFile", domain.Name)
		}

		if domain.KeyFile == "" {
			return fmt.Errorf("domain %s must have a keyFile", domain.Name)
		}

		return nil
	}

	processDomains := func() error {

		var errs *multierror.Error

		for _, domain := range config.Domains {
			err := processDomain(domain)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		return errs.ErrorOrNil()
	}

	processSynologyDomains := func() error {

		domainsLen := len(config.Domains)

		if domainsLen >= 0 {
			return fmt.Errorf("Missing Domain")
		}

		if domainsLen > 1 {
			return fmt.Errorf("There are %d domains in the configuration and there should only be 1", domainsLen)
		}

		domain := config.Domains[0]

		if domain.Name == "" {
			return fmt.Errorf("domain must have a name")
		}

		if domain.CertFile != "" {
			zap.L().Debug(fmt.Sprintf("Domain %s for synology client has certFile set; it will be ignored", domain.Name))
			domain.CertFile = ""
		}

		if domain.KeyFile != "" {
			zap.L().Debug(fmt.Sprintf("Domain %s for synology client has keyFile set; it will be ignored", domain.Name))
			domain.KeyFile = ""
		}

		if domain.FullChain != "" {
			zap.L().Debug(fmt.Sprintf("Domain %s for synology client has fullChain set; it will be ignored", domain.Name))
			domain.FullChain = ""
		}

		return nil
	}

	osType := getOS()

	if config.IgnoreOSType {
		osType = OSTypeEmpty
		zap.L().Debug("IgnoreOSType is set to true")
	} else {
		zap.L().Debug(fmt.Sprintf("OSType is %s", string(osType)))
	}

	switch osType {

	case OSTypeSynology:
		err := processSynologyDomains()
		if err != nil {
			return nil, err
		}

	default:
		err := processDomains()
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		osType: osType,
		config: config,
		client: libclient.New(&libclient.Config{
			Secret:     config.Secret,
			Server:     config.Server,
			SkipVerify: config.SkipVerify,
		}),
	}, nil
}

func (t *Client) Run(ctx context.Context) error {

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
			return fmt.Errorf("error creating cert directory %s; %w", filepath.Dir(domain.CertFile), err)
		}

		err = os.WriteFile(domain.CertFile, cert.GetCertPEM(), filePerm)
		if err != nil {
			return fmt.Errorf("error writing cert file %s; %w", domain.CertFile, err)
		}

		err = os.MkdirAll(filepath.Dir(domain.KeyFile), dirPerm)
		if err != nil {
			return err
		}

		err = os.WriteFile(domain.KeyFile, cert.GetKeyPEM(), filePerm)
		if err != nil {
			return fmt.Errorf("error writing key file %s; %w", domain.KeyFile, err)
		}

		if domain.FullChain != "" {

			err := os.MkdirAll(filepath.Dir(domain.CertFile), dirPerm)
			if err != nil {
				return fmt.Errorf("error creating fullChain directory %s; %w", filepath.Dir(domain.FullChain), err)
			}

			err = os.WriteFile(domain.FullChain, cert.GetCertPEM(), filePerm)
			if err != nil {
				return fmt.Errorf("error writing fullChain file %s; %w", domain.FullChain, err)
			}
		}

		if domain.Hook == nil {
			zap.L().Debug(fmt.Sprintf("Hook config not present for domain %s", domain.Name))
			return nil
		}

		zap.L().Debug(fmt.Sprintf("Running Hook for domain %s", domain.Name))
		cmd := exec.Command(domain.Hook.Name, domain.Hook.Args...)
		rawoutput, err := cmd.CombinedOutput()
		output := strings.ReplaceAll(string(rawoutput), "\n", "")

		if err != nil {
			return fmt.Errorf("Command '%s' for Domain %s returned output %s and error %w", domain.Hook.GetCmd(), domain.Name, output, err)
		}

		if logger.Trace {
			zap.L().Debug(fmt.Sprintf("Command '%s' for Domain %s returned output %s", domain.Hook.GetCmd(), domain.Name, output))
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

		if domain == nil {
			panic("this should not happen")
		}

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

	switch t.osType {

	case OSTypeSynology:
		return runSynologyMode()

	}

	if t.config.Daemon {
		runTick()
		runDaemon()
		return nil
	}
	return run()

}
