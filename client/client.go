package client

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	logger "github.com/jodydadescott/jody-go-logger"
	"go.uber.org/zap"

	"github.com/jodydadescott/home-simplecert/libclient"
	"github.com/jodydadescott/home-simplecert/types"
	"github.com/jodydadescott/home-simplecert/util"
)

type Client struct {
	config *Config
	client *libclient.Client
	osType OSType
}

func New(config *Config) (*Client, error) {

	if config == nil {
		return nil, fmt.Errorf("config is required")
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
			return fmt.Errorf("domain must have a Name")
		}

		if domain.DomainName == "" {
			return fmt.Errorf("domain must have a DomainName")
		}

		if domain.CertFile == "" && domain.FullChain == "" {
			return fmt.Errorf("domain %s: one or both of the following is required: CertFile, FullChain", domain.Name)
		}

		if domain.KeyFile == "" {
			return fmt.Errorf("domain %s: KeyFile is required", domain.Name)
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

		if domainsLen <= 0 {
			return fmt.Errorf("missing Domain")
		}

		if domainsLen > 1 {
			return fmt.Errorf("there are %d domains in the configuration and there should only be 1", domainsLen)
		}

		domain := config.Domains[0]

		if domain.DomainName == "" {
			return fmt.Errorf("domain must have a DomainName")
		}

		if domain.Name != "" {
			zap.L().Debug(fmt.Sprintf("Domain Synology client has Name %s; it will be ignored", domain.Name))
		}

		if domain.CertFile != "" {
			zap.L().Debug(fmt.Sprintf("Domain Synology client has CertFile %s; it will be ignored", domain.CertFile))
		}

		if domain.KeyFile != "" {
			zap.L().Debug(fmt.Sprintf("Domain Synology client has KeyFile %s; it will be ignored", domain.KeyFile))
		}

		if domain.FullChain != "" {
			zap.L().Debug(fmt.Sprintf("Domain Synology client has FullChain %s; it will be ignored", domain.FullChain))
		}

		if domain.Hook != nil {
			zap.L().Debug(fmt.Sprintf("Domain %s: Synology client has a Hook; it will be ignored", domain.Name))
			zap.L().Debug("Domain Synology client has Hook; it will be ignored")
		}

		b, err := os.ReadFile(SynologyDefaultFile)
		if err != nil {
			return err
		}

		domain.Name = Synology

		cerFile := strings.ReplaceAll(string(b), "\n", "")
		certDir := filepath.Join(filepath.Dir(SynologyDefaultFile), cerFile)

		domain.Hook = &Hook{
			Name: "systemctl",
		}
		domain.Hook.AddArgs("restart", "nginx.service")

		domain.CertFile = filepath.Join(certDir, SynologyCertFile)
		domain.KeyFile = filepath.Join(certDir, SynologyKeyFile)
		domain.FullChain = filepath.Join(certDir, SynologyChainFile)

		return nil
	}

	processUnifiDomains := func() error {

		domainsLen := len(config.Domains)

		if domainsLen <= 0 {
			return fmt.Errorf("missing Domain")
		}

		if domainsLen > 1 {
			return fmt.Errorf("there are %d domains in the configuration and there should only be 1", domainsLen)
		}

		domain := config.Domains[0]

		if domain.DomainName == "" {
			return fmt.Errorf("domain must have a DomainName")
		}

		if domain.Name != "" {
			zap.L().Debug(fmt.Sprintf("Domain Unifi client has Name %s; it will be ignored", domain.Name))
		}

		if domain.CertFile != "" {
			zap.L().Debug(fmt.Sprintf("Domain Unifi client has CertFile %s; it will be ignored", domain.CertFile))
		}

		if domain.KeyFile != "" {
			zap.L().Debug(fmt.Sprintf("Domain Unifi client has KeyFile %s; it will be ignored", domain.KeyFile))
		}

		if domain.FullChain != "" {
			zap.L().Debug(fmt.Sprintf("Domain Unifi client has FullChain %s; it will be ignored", domain.FullChain))
		}

		if domain.Hook != nil {
			zap.L().Debug(fmt.Sprintf("Domain %s: Unifi client has a Hook; it will be ignored", domain.Name))
			zap.L().Debug("Domain Unifi client has Hook; it will be ignored")
		}

		domain.Name = Unifi

		domain.CertFile = UnifiCertFile
		domain.KeyFile = UnifiKeyFile

		domain.Hook = &Hook{
			Name: "systemctl",
		}
		domain.Hook.AddArgs("restart", "unifi-core.service")

		return nil
	}

	getOS := func() OSType {

		osType := OSTypeFromString(runtime.GOOS)

		if osType == OSTypeLinux {
			if util.FileExist(DetectSynologyFile) {
				osType = OSTypeSynology
			}
		}

		if osType == OSTypeLinux {
			if util.FileExist(DetectSynologyFile) {
				return OSTypeSynology
			}

			if util.FileExist(DetectUnifiFile) {
				return OSTypeUnifi
			}

		}

		return osType
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
		if config.Daemon {
			zap.L().Debug("Daemon is set to true on OSType Synology; it will be ignored")
			config.Daemon = false
		}

	case OSTypeUnifi:
		err := processUnifiDomains()
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

	execCmd := func(name string, args []string) error {

		logname := name
		for _, s := range args {
			logname = logname + " " + s
		}

		if logger.Trace {
			zap.L().Debug(fmt.Sprintf("Executing cmd %s", logname))
		}

		output, err := util.ExecuteCommand(name, args...)

		if err == nil {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("cmd %s returned output: ", output))
			}
			return nil
		}

		return fmt.Errorf("executing cmd %s returned error: %w", logname, err)
	}

	compare := func(file string, abytes []byte) bool {

		if !util.FileExist(file) {
			return false
		}

		bbytes, err := os.ReadFile(file)
		if err != nil {
			return false
		}

		return bytes.Equal(abytes, bbytes)
	}

	writeFile := func(name string, data []byte) error {
		err := os.MkdirAll(filepath.Dir(name), types.DirPerm)
		if err != nil {
			return fmt.Errorf("error creating cert directory %s; %w", filepath.Dir(name), err)
		}

		err = os.WriteFile(name, data, types.FilePerm)
		if err != nil {
			return fmt.Errorf("error writing %s; %w", name, err)
		}
		return nil
	}

	process := func(domain *Domain, cert *CR) error {

		result := false

		if domain.KeyFile != "" {

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: has KeyFile %s", domain.Name, domain.KeyFile))
			}

			data := cert.GetKeyPEM()

			if !compare(domain.KeyFile, data) {

				result = true
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: KeyFile %s changed", domain.Name, domain.KeyFile))
				}

				err := writeFile(domain.KeyFile, data)
				if err != nil {
					return err
				}

			} else {
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: KeyFile %s unchanged", domain.Name, domain.KeyFile))
				}
			}
		} else {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: does not have KeyFile", domain.Name))
			}
		}

		if domain.CertFile != "" {

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: has CertFile %s", domain.Name, domain.CertFile))
			}

			data := cert.GetCertPEM()

			if !compare(domain.CertFile, data) {

				result = true
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: CertFile %s changed", domain.Name, domain.CertFile))
				}

				err := writeFile(domain.CertFile, data)
				if err != nil {
					return err
				}

			} else {
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: CertFile %s unchanged", domain.Name, domain.CertFile))
				}
			}
		} else {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: does not have CertFile", domain.Name))
			}
		}

		if domain.FullChain != "" {

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: has CertFile %s", domain.Name, domain.FullChain))
			}

			data := cert.GetCertPEM()

			if !compare(domain.FullChain, data) {

				result = true
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: FullChain %s changed", domain.Name, domain.FullChain))
				}

				err := writeFile(domain.FullChain, data)
				if err != nil {
					return err
				}

			} else {
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: FullChain %s unchanged", domain.Name, domain.FullChain))
				}
			}
		} else {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: does not have FullChain", domain.Name))
			}
		}

		if result {

			if domain.Hook == nil {
				zap.L().Info(fmt.Sprintf("Domain %s: changed; no hook configured", domain.Name))
				return nil
			}

			zap.L().Info(fmt.Sprintf("Domain %s: changed; executing hook", domain.Name))

			return execCmd(domain.Hook.Name, domain.Hook.Args)
		}

		zap.L().Info(fmt.Sprintf("Domain %s: unchanged", domain.Name))

		return nil
	}

	run := func() error {

		zap.L().Debug("Processing domains")

		var errs *multierror.Error

		for _, domain := range t.config.Domains {

			cert, err := t.client.GetCert(domain.DomainName)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
				continue
			}

			err = process(domain, cert)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("Domain %s %w", domain.Name, err))
				continue
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
			refreshInterval = DefaultRefreshInterval
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

	zap.L().Debug("Client is now running")

	defer func() {
		t.client.Shutdown()
		zap.L().Debug("Client is shutting down")
	}()

	if t.config.Daemon {
		runTick()
		runDaemon()
		return nil
	}

	return run()

}
