package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/hokaccha/go-prettyjson"
	logger "github.com/jodydadescott/jody-go-logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/jodydadescott/home-simplecert/client"
	"github.com/jodydadescott/home-simplecert/server"
	"github.com/jodydadescott/home-simplecert/types"
	"github.com/jodydadescott/home-simplecert/util"
)

func getExampleConfig() *Config {
	return &Config{
		Notes:  ConfigNotes,
		Server: server.ExampleConfig(),
		Client: client.ExampleConfig(),
		Logger: &Logger{
			LogLevel: logger.DebugLevel,
		},
	}
}

func getExampleConfigCmd() *cobra.Command {

	getConfigCmd := func(format string) *cobra.Command {

		upperFormat := strings.ToUpper(format)
		lowerFormat := strings.ToLower(format)

		config := getExampleConfig()

		return &cobra.Command{
			Use:  lowerFormat,
			Long: fmt.Sprintf("generates new example config in %s format", upperFormat),
			RunE: func(cmd *cobra.Command, args []string) error {

				var o []byte

				switch cmd.Use {

				case "json":
					o, _ = json.Marshal(config)

				case "yaml":
					o, _ = yaml.Marshal(config)

				case "pretty-json":
					o, _ = prettyjson.Marshal(config)

				default:
					return fmt.Errorf("supported formats are json, yaml, and pretty-json")

				}

				fmt.Print(string(o))
				return nil
			},
		}
	}

	cmd := &cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(getConfigCmd("json"), getConfigCmd("yaml"), getConfigCmd("pretty-json"))

	return cmd
}

var (
	configFileArg string
	debugLevelArg string

	rootCmd = &cobra.Command{
		Use: BinaryName,
	}

	versionCmd = &cobra.Command{
		Use:  "version",
		Long: "Returns the version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(types.CodeVersion)
			return nil
		},
	}

	installCmd = &cobra.Command{
		Use:  "install",
		Long: "installs the binary, config and init files",
		RunE: func(cmd *cobra.Command, args []string) error {

			installSystemd := func() error {

				binaryFilePath, err := os.Executable()
				if err != nil {
					return err
				}

				if !util.FileExist(SystemdServiceFile) {
					err = os.WriteFile(SystemdServiceFile, []byte(systemD()), types.SecureFilePerm)
					if err != nil {
						return err
					}
				}

				if !util.FileExist(DefaultConfigFile) {

					config := getExampleConfig()
					o, _ := yaml.Marshal(config)

					err = os.WriteFile(DefaultConfigFile, o, types.FilePerm)
					if err != nil {
						return err
					}

					fmt.Fprintf(os.Stderr, "Config file %s was created", DefaultConfigFile)
				}

				binary, err := os.ReadFile(binaryFilePath)
				if err != nil {
					return err
				}

				err = os.MkdirAll(BinaryInstallPath, types.DirPerm)
				if err != nil {
					return err
				}

				err = os.WriteFile(filepath.Join(BinaryInstallPath, BinaryName), binary, types.ExePerm)
				if err != nil {
					return err
				}

				fmt.Fprintf(os.Stderr, "Installed as a systemd service. Configure the file %s and then run the commands:\n", DefaultConfigFile)
				fmt.Fprintf(os.Stderr, "systemctl enable home-simplecert && systemctl start home-simplecert\n")
				return nil

			}

			switch runtime.GOOS {

			case "linux":
				if util.FileExist("/etc/systemd/system") {
					return installSystemd()
				}

			default:
				return fmt.Errorf("install not supported on OS %s", runtime.GOOS)

			}

			return nil
		},
	}

	runCmd = &cobra.Command{

		Use: "run",

		RunE: func(cmd *cobra.Command, args []string) error {

			getConfig := func(configFile string) (*Config, error) {

				if !util.FileExist(configFile) {
					return nil, fmt.Errorf("config file %s does not exist", configFile)
				}

				fileStats, err := os.Stat(configFile)
				if err != nil {
					return nil, err
				}

				permissions := fileStats.Mode().Perm()
				if permissions != types.SecureFilePerm {
					return nil, fmt.Errorf("config file %s has overly promiscuous permissions", configFile)
				}

				content, err := os.ReadFile(configFile)
				if err != nil {
					return nil, err
				}

				var config Config
				err = json.Unmarshal(content, &config)
				if err == nil {
					return &config, nil
				}

				var errs *multierror.Error

				errs = multierror.Append(errs, err)

				err = yaml.Unmarshal(content, &config)
				if err == nil {
					return &config, nil
				}

				errs = multierror.Append(errs, err)

				return nil, errs.ErrorOrNil()
			}

			errc := make(chan error, 2)

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			interruptChan := make(chan os.Signal, 1)
			signal.Notify(interruptChan, os.Interrupt)

			configFile := configFileArg

			var config *Config

			if configFile == "" {
				configFile = os.Getenv(ConfigEnvVar)
			}

			if configFile == "" {
				configFile = DefaultConfigFile
			}

			config, err := getConfig(configFile)
			if err != nil {
				return err
			}

			debugLevel := debugLevelArg
			if debugLevel == "" {
				debugLevel = os.Getenv(DebugEnvVar)
			}
			if debugLevel != "" {
				if config.Logger == nil {
					config.Logger = &logger.Config{}
				}
				err := config.Logger.ParseLogLevel(debugLevel)
				if err != nil {
					return err
				}
			}

			if config.Logger != nil {
				logger.SetConfig(config.Logger)
			}

			var clientRunner *client.Client
			var serverRunner *server.Server

			if config.Client != nil {
				x, err := client.New(config.Client)
				if err != nil {
					return err
				}
				clientRunner = x
			}

			if config.Server != nil {
				x, err := server.New(config.Server)
				if serverRunner != nil {
					return err
				}
				serverRunner = x
			}

			var wg sync.WaitGroup

			if clientRunner != nil {
				zap.L().Debug("Running Client")
				wg.Add(1)
				go func() {
					ctx, cancel = context.WithCancel(cmd.Context())
					defer func() {
						cancel()
						wg.Done()
					}()
					errc <- clientRunner.Run(ctx)
				}()
			}

			if serverRunner != nil {
				zap.L().Debug("Running Server")
				wg.Add(1)
				go func() {
					ctx, cancel = context.WithCancel(cmd.Context())
					defer func() {
						cancel()
						wg.Done()
					}()
					errc <- serverRunner.Run(ctx)
				}()
			}

			select {

			case <-interruptChan: // first signal, cancel context
				cancel()

			case <-ctx.Done():

			case err := <-errc:
				return err

			}

			wg.Wait()

			return nil
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {

	configCmd := getExampleConfigCmd()

	rootCmd.AddCommand(versionCmd, configCmd, runCmd, installCmd)
	runCmd.PersistentFlags().StringVarP(&configFileArg, "config", "c", "", fmt.Sprintf("config file; env var is %s", ConfigEnvVar))
	runCmd.PersistentFlags().StringVarP(&debugLevelArg, "debug", "D", "", fmt.Sprintf("debug level (TRACE, DEBUG, INFO, WARN, ERROR) to STDERR; env var is %s", ConfigEnvVar))
}
