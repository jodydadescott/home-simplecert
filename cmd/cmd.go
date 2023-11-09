package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hokaccha/go-prettyjson"
	logger "github.com/jodydadescott/jody-go-logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/jodydadescott/home-simplecert/client"
	"github.com/jodydadescott/home-simplecert/server"
)

func getExampleConfigCmd(use string, config *Config) *cobra.Command {

	getConfigCmd := func(format string) *cobra.Command {

		upperFormat := strings.ToUpper(format)
		lowerFormat := strings.ToLower(format)

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
					return fmt.Errorf("Supported formats are json, yaml, and pretty-json")

				}

				fmt.Print(string(o))
				return nil
			},
		}
	}

	cmd := &cobra.Command{
		Use: use,
	}

	cmd.AddCommand(getConfigCmd("json"), getConfigCmd("yaml"), getConfigCmd("pretty-json"))

	return cmd
}

func getConfigCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(getExampleConfigCmd("client", ExampleClientConfig()), getExampleConfigCmd("server", ExampleServerConfig()))

	return cmd
}

func getClientConfig(configFile string) (*ClientConfig, error) {

	var errs *multierror.Error

	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config ClientConfig
	err = json.Unmarshal(content, &config)
	if err == nil {
		return &config, nil
	}

	errs = multierror.Append(errs, err)

	err = yaml.Unmarshal(content, &config)
	if err == nil {
		return &config, nil
	}

	errs = multierror.Append(errs, err)

	return nil, errs.ErrorOrNil()
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
			fmt.Println(CodeVersion)
			return nil
		},
	}

	runCmd = &cobra.Command{

		Use: "run",

		RunE: func(cmd *cobra.Command, args []string) error {

			getConfig := func(configFile string) (*Config, error) {

				var errs *multierror.Error

				content, err := os.ReadFile(configFile)
				if err != nil {
					return nil, err
				}

				var config Config
				err = json.Unmarshal(content, &config)
				if err == nil {
					return &config, nil
				}

				errs = multierror.Append(errs, err)

				err = yaml.Unmarshal(content, &config)
				if err == nil {
					return &config, nil
				}

				errs = multierror.Append(errs, err)

				return nil, errs.ErrorOrNil()
			}

			run := func(runner Runner) error {

				ctx, cancel := context.WithCancel(cmd.Context())

				interruptChan := make(chan os.Signal, 1)
				signal.Notify(interruptChan, os.Interrupt)

				go func() {
					select {
					case <-interruptChan: // first signal, cancel context
						cancel()
					case <-ctx.Done():
					}
					<-interruptChan // second signal, hard exit
				}()

				return runner.Run(ctx)
			}

			configFile := configFileArg

			var config *Config

			if configFile == "" {
				configFile = os.Getenv(ConfigEnvVar)
			}

			if configFile != "" {
				tmpConfig, err := getConfig(configFileArg)
				if err != nil {
					return err
				}
				config = tmpConfig.Clone()
			} else {
				return fmt.Errorf("Config is required")
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

			if config.Client != nil && config.Server != nil {
				return fmt.Errorf("Config contains both client and server. It should only contain one or the other")
			}

			if config.Client != nil {

				zap.L().Debug("Running Client")
				runner, err := client.New(config.Client)
				if err != nil {
					return err
				}
				return run(runner)
			}

			zap.L().Debug("Running Server")
			runner, err := server.New(config.Server)
			if err != nil {
				return err
			}
			return run(runner)

		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd, getConfigCmd(), runCmd)
	runCmd.PersistentFlags().StringVarP(&configFileArg, "config", "c", "", fmt.Sprintf("config file; env var is %s", ConfigEnvVar))
	runCmd.PersistentFlags().StringVarP(&debugLevelArg, "debug", "d", "", fmt.Sprintf("debug level (TRACE, DEBUG, INFO, WARN, ERROR) to STDERR; env var is %s", ConfigEnvVar))
}
