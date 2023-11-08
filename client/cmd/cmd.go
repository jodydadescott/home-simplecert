package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/hashicorp/go-multierror"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/jodydadescott/home-simplecert/client/client"
)

type Config = client.Config

const (
	BinaryName   = "home-simplecert-client"
	CodeVersion  = "0.1.0"
	ConfigEnvVar = "CONFIG"
	DebugEnvVar  = "DEBUG"
)

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

	configCmd = &cobra.Command{
		Use:  "config",
		Long: "generates new example config",
	}

	jsonConfigCmd = &cobra.Command{
		Use:  "json",
		Long: "generates new example config in JSON format",
		RunE: func(cmd *cobra.Command, args []string) error {
			o, _ := json.Marshal(client.ExampleConfig())
			fmt.Print(string(o))
			return nil
		},
	}

	jsonConfigPrettyCmd = &cobra.Command{
		Use:  "pretty-json",
		Long: "generates new example config in pretty JSON format",
		RunE: func(cmd *cobra.Command, args []string) error {
			o, _ := prettyjson.Marshal(client.ExampleConfig())
			fmt.Print(string(o))
			return nil
		},
	}

	yamlConfigCmd = &cobra.Command{
		Use:  "yaml",
		Long: "generates new example config in YAML format",
		RunE: func(cmd *cobra.Command, args []string) error {
			o, _ := yaml.Marshal(client.ExampleConfig())
			fmt.Print(string(o))
			return nil
		},
	}

	runCmd = &cobra.Command{

		Use: "run",

		RunE: func(cmd *cobra.Command, args []string) error {

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
				config = &Config{}
			}

			debugLevel := ""
			if debugLevelArg != "" {
				debugLevel = debugLevelArg
			} else {
				if os.Getenv(DebugEnvVar) != "" {
					debugLevel = os.Getenv(DebugEnvVar)
				}
			}

			if debugLevel != "" {
				config.Logger.ParseLogLevel(debugLevel)
			}

			c, err := client.New(config)
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

			return c.Run(ctx)
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
	runCmd.PersistentFlags().StringVarP(&debugLevelArg, "debug", "d", "", fmt.Sprintf("debug level (TRACE, DEBUG, INFO, WARN, ERROR) to STDERR; env var is %s", ConfigEnvVar))
	configCmd.AddCommand(jsonConfigCmd, jsonConfigPrettyCmd, yamlConfigCmd)
	rootCmd.AddCommand(versionCmd, configCmd, runCmd)
}
