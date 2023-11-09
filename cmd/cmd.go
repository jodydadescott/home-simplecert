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
	"gopkg.in/yaml.v2"

	"github.com/jodydadescott/home-simplecert/client"
	"github.com/jodydadescott/home-simplecert/server"
)

type ClientConfig = client.Config
type ServerConfig = server.Config

const (
	BinaryName   = "home-simplecert"
	CodeVersion  = "0.1.0"
	ConfigEnvVar = "CONFIG"
	DebugEnvVar  = "DEBUG"
)

func getClientExampleConfigCmd(use string, config *ClientConfig) *cobra.Command {

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

func getServerExampleConfigCmd() *cobra.Command {

	config := server.ExampleConfig()

	getConfigCmd := func(format string) *cobra.Command {

		upperFormat := strings.ToUpper(format)
		lowerFormat := strings.ToUpper(format)

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
		Use: "server",
	}

	cmd.AddCommand(getConfigCmd("json"), getConfigCmd("yaml"), getConfigCmd("pretty-json"))

	return cmd
}

func getClientConfigCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use: "client",
	}

	cmd.AddCommand(getClientExampleConfigCmd("synology", client.ExampleSynologyConfig()), getClientExampleConfigCmd("normal", client.ExampleNormalConfig()))
	return cmd
}

func getConfigCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(getClientConfigCmd(), getServerExampleConfigCmd())
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

func getServerConfig(configFile string) (*ServerConfig, error) {

	var errs *multierror.Error

	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
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
	}

	clientRunCmd = &cobra.Command{

		Use: "client",

		RunE: func(cmd *cobra.Command, args []string) error {

			configFile := configFileArg

			var config *ClientConfig

			if configFile == "" {
				configFile = os.Getenv(ConfigEnvVar)
			}

			if configFile != "" {
				tmpConfig, err := getClientConfig(configFileArg)
				if err != nil {
					return err
				}
				config = tmpConfig.Clone()
			} else {
				config = &ClientConfig{}
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

	serverRunCmd = &cobra.Command{

		Use: "server",

		RunE: func(cmd *cobra.Command, args []string) error {

			configFile := configFileArg

			if configFile == "" {
				configFile = os.Getenv(ConfigEnvVar)
			}

			if configFile == "" {
				return fmt.Errorf("configFile is required; set using option or env var %s", ConfigEnvVar)
			}

			config, err := getServerConfig(configFileArg)
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

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	runCmd.AddCommand(clientRunCmd, serverRunCmd)
	rootCmd.AddCommand(versionCmd, getConfigCmd(), runCmd)
	runCmd.PersistentFlags().StringVarP(&configFileArg, "config", "c", "", fmt.Sprintf("config file; env var is %s", ConfigEnvVar))
	runCmd.PersistentFlags().StringVarP(&debugLevelArg, "debug", "d", "", fmt.Sprintf("debug level (TRACE, DEBUG, INFO, WARN, ERROR) to STDERR; env var is %s", ConfigEnvVar))
}
