package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"gopkg.in/yaml.v2"

	"github.com/hashicorp/go-multierror"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"

	"github.com/jodydadescott/home-simplecert/server/server"
	logger "github.com/jodydadescott/jody-go-logger"
)

const (
	BinaryName   = "home-simplecert-server"
	ConfigEnvVar = "CONFIG"
	DebugEnvVar  = "DEBUG"
	CodeVersion  = "0.1.0"
)

type Config = server.Config

var (
	configFileArg string
	debugLevelArg string

	rootCmd = &cobra.Command{
		Use: BinaryName,
	}

	versionCmd = &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(CodeVersion)
			return nil
		},
	}

	generateConfigCmd = &cobra.Command{
		Use: "generate-config",
	}

	generateJsonConfigCmd = &cobra.Command{
		Use: "json",
		RunE: func(cmd *cobra.Command, args []string) error {
			o, _ := json.Marshal(server.ExampleConfig())
			fmt.Println(string(o))
			return nil
		},
	}

	generateJsonConfigPrettyCmd = &cobra.Command{
		Use: "pretty-json",
		RunE: func(cmd *cobra.Command, args []string) error {
			o, _ := prettyjson.Marshal(server.ExampleConfig())
			fmt.Println(string(o))
			return nil
		},
	}

	generateYamlConfigCmd = &cobra.Command{
		Use: "yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			o, _ := yaml.Marshal(server.ExampleConfig())
			fmt.Println(string(o))
			return nil
		},
	}

	runCmd = &cobra.Command{

		Use: "run",

		RunE: func(cmd *cobra.Command, args []string) error {

			configFile := configFileArg

			if configFile == "" {
				configFile = os.Getenv(ConfigEnvVar)
			}

			if configFile == "" {
				return fmt.Errorf("configFile is required; set using option or env var %s", ConfigEnvVar)
			}

			config, err := getConfig(configFileArg)
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

			s, err := server.New(config)
			if err != nil {
				return err
			}

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

			return s.Run(ctx)

		},
	}
)

func getConfig(configFile string) (*Config, error) {

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

func Execute() error {
	return rootCmd.Execute()
}

func init() {

	runCmd.PersistentFlags().StringVarP(&configFileArg, "config", "c", "", fmt.Sprintf("config file; env var is %s", ConfigEnvVar))
	runCmd.PersistentFlags().StringVarP(&debugLevelArg, "debug", "d", "", fmt.Sprintf("debug level (WIRE, TRACE, DEBUG, INFO, WARN, ERROR) to STDERR; env var is %s", ConfigEnvVar))

	generateConfigCmd.AddCommand(generateJsonConfigCmd, generateJsonConfigPrettyCmd, generateYamlConfigCmd)
	rootCmd.AddCommand(versionCmd, generateConfigCmd, runCmd)
}
