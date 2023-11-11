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
			return fmt.Errorf("each domain must have a name")
		}

		if domain.CertFile == "" && domain.FullChain == "" && domain.Keystore == nil {
			return fmt.Errorf("Domain %s: one or more of the following is required: CertFile, FullChain, Keystore", domain.Name)
		}

		if domain.Keystore != nil {
			if domain.Keystore.File == "" && domain.Keystore.Secret == "" {
				return fmt.Errorf("Domain %s: if Keystore is set then both File and Secret must be set", domain.Name)
			}
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
			zap.L().Debug(fmt.Sprintf("Domain %s: Synology client has CertFile set; it will be ignored", domain.Name))
			domain.CertFile = ""
		}

		if domain.KeyFile != "" {
			zap.L().Debug(fmt.Sprintf("Domain %s: Synology client has KeyFile set; it will be ignored", domain.Name))
			domain.KeyFile = ""
		}

		if domain.FullChain != "" {
			zap.L().Debug(fmt.Sprintf("Domain %s: Synology client has FullChain set; it will be ignored", domain.Name))
			domain.FullChain = ""
		}

		if domain.Keystore != nil {
			zap.L().Debug(fmt.Sprintf("Domain %s: Synology client has Keystore set; it will be ignored", domain.Name))
			domain.Keystore = nil
		}

		if domain.Hook != nil {
			zap.L().Debug(fmt.Sprintf("Domain %s: Synology client has a Hook; it will be ignored", domain.Name))
			domain.Hook = nil
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

	execCmd := func(name string, args []string) error {

		logname := name
		for _, s := range args {
			logname = logname + " " + s
		}

		if logger.Trace {
			zap.L().Debug(fmt.Sprintf("Executing cmd %s", logname))
		}

		cmd := exec.Command(name, args...)
		rawoutput, err := cmd.CombinedOutput()
		output := strings.ReplaceAll(string(rawoutput), "\n", "")

		if err == nil {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("cmd %s returned output: ", output))
			}
			return nil
		}

		return fmt.Errorf("Executing cmd %s returned error: %w", logname, err)
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

	_tmpdir := ""
	defer func() {
		if _tmpdir != "" {
			os.RemoveAll(_tmpdir)
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("removed tmp dir %s", _tmpdir))
			}
		}
	}()

	tmpdir := func() string {
		if _tmpdir == "" {
			_tmpdir = os.TempDir()

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("created tmp dir %s", _tmpdir))
			}
		}
		return _tmpdir
	}

	makeKeystore := func(pass string, data []byte) ([]byte, error) {

		// openssl pkcs12 -export -out cache/fullchain.pkcs12 -in cache/fullchain  -passout pass:openhab

		tmp := tmpdir()

		inputFile := filepath.Join(tmp, "pem_input")
		outputFile := filepath.Join(tmp, "pkcs12_output")

		err := os.WriteFile(inputFile, data, filePerm)
		if err != nil {
			return nil, fmt.Errorf("makeKeystore error: %w", err)
		}

		args := []string{"pkcs12", "-export", "-out", outputFile, "-in", inputFile, "-passout", fmt.Sprintf("pass:%s", pass)}

		err = execCmd("openssl", args)
		if err != nil {
			return nil, err
		}

		result, err := os.ReadFile(outputFile)
		if err != nil {
			return nil, fmt.Errorf("makeKeystore error: %w", err)
		}

		return result, nil
	}

	writeFile := func(name string, data []byte) error {
		err := os.MkdirAll(filepath.Dir(name), dirPerm)
		if err != nil {
			return fmt.Errorf("error creating cert directory %s; %w", filepath.Dir(name), err)
		}

		err = os.WriteFile(name, data, filePerm)
		if err != nil {
			return fmt.Errorf("error writing %s; %w", name, err)
		}
		return nil
	}

	process := func(domain *Domain, cert *CR) error {

		result := false

		if domain.KeyFile != "" {

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: has KeyFile", domain.Name))
			}

			data := cert.GetKeyPEM()

			if !compare(domain.KeyFile, data) {

				result = true
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: KeyFile changed", domain.Name))
				}

				err := writeFile(domain.KeyFile, data)
				if err != nil {
					return err
				}

			} else {
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: KeyFile unchanged", domain.Name))
				}
			}
		} else {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: does not have KeyFile", domain.Name))
			}
		}

		if domain.CertFile != "" {

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: has CertFile", domain.Name))
			}

			data := cert.GetCertPEM()

			if !compare(domain.CertFile, data) {

				result = true
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: CertFile changed", domain.Name))
				}

				err := writeFile(domain.CertFile, data)
				if err != nil {
					return err
				}

			} else {
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: CertFile unchanged", domain.Name))
				}
			}
		} else {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: does not have CertFile", domain.Name))
			}
		}

		if domain.FullChain != "" {

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: has FullChain", domain.Name))
			}

			data := cert.GetCertPEM()

			if !compare(domain.FullChain, data) {

				result = true
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: FullChain changed", domain.Name))
				}

				err := writeFile(domain.FullChain, data)
				if err != nil {
					return err
				}

			} else {
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: FullChain unchanged", domain.Name))
				}
			}
		} else {
			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: does not have FullChain", domain.Name))
			}
		}

		if domain.Keystore != nil {

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: has Keystore", domain.Name))
			}

			data := cert.GetCertPEM()
			data = append(data, cert.GetKeyPEM()...)

			keystore, err := makeKeystore(domain.Keystore.Secret, data)
			if err != nil {
				return err
			}

			if !compare(domain.Keystore.File, keystore) {

				result = true
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: Keystore changed", domain.Name))
				}

				err := writeFile(domain.Keystore.File, keystore)
				if err != nil {
					return err
				}

			} else {
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: Keystore unchanged", domain.Name))
				}
			}
		}

		if result {

			if logger.Trace {
				zap.L().Debug(fmt.Sprintf("Domain %s: changed", domain.Name))
			}

			if domain.Hook == nil {
				if logger.Trace {
					zap.L().Debug(fmt.Sprintf("Domain %s: no hook present", domain.Name))
				}
				return nil
			}
			zap.L().Debug(fmt.Sprintf("Domain %s: calling hook", domain.Name))
			return execCmd(domain.Hook.Name, domain.Hook.Args)
		}

		if logger.Trace {
			zap.L().Debug(fmt.Sprintf("Domain %s: unchanged", domain.Name))
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

		return process(domain, cert)
	}

	defer func() {
		t.client.Shutdown()
		zap.L().Debug("Shutting down Client")
	}()

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
