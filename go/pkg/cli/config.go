// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

const (
	addressKey = "service-address"

	tlsCertPathKey = "tls.certPath"
	tlsKeyPathKey  = "tls.keyPath"
	noTLSKey       = "no-tls"
	authHeaderKey  = "auth-header"

	addressFlag     = "service-address"
	tlsCertPathFlag = "tls-cert-path"
	tlsKeyPathFlag  = "tls-key-path"
	noTLSFlag       = "no-tls"
	// AuthHeaderFlag - the flag name
	AuthHeaderFlag = "auth-header"

	// Authorization the header keyword
	Authorization = "authorization"
)

var configDir = ".onos"

var configName string

var configOptions = []string{
	addressKey,
	tlsCertPathKey,
	tlsKeyPathKey,
	noTLSKey,
	authHeaderKey,
}

// SetConfigDir sets the name of the config directory as a relative path under the home directory where the config
// file will be created/expected.
func SetConfigDir(relativePath string) {
	configDir = relativePath
}

// CreateConfig creates a CLI config file including its parent directory if necessary. It is idempotent.
func CreateConfig(verbose bool) error {
	if err := viper.ReadInConfig(); err == nil {
		return nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(home+"/"+configDir, 0777); err != nil {
		return err
	}

	filePath := home + "/" + configDir + "/" + configName + ".yaml"
	f, err := os.Create(filePath)

	if err != nil {
		return err
	}
	_ = f.Close()

	if err = viper.WriteConfig(); err != nil {
		return err
	}
	if verbose {
		_, _ = fmt.Fprintf(GetOutput(), "Created %s\n", filePath)
	}
	return nil
}

// AddConfigFlags :
func AddConfigFlags(cmd *cobra.Command, serviceAddress string) {
	viper.SetDefault(addressKey, serviceAddress)

	cmd.PersistentFlags().String(addressFlag, viper.GetString(addressKey), "the gRPC endpoint")
	cmd.PersistentFlags().String(tlsCertPathFlag, viper.GetString(tlsCertPathKey), "the path to the TLS certificate")
	cmd.PersistentFlags().String(tlsKeyPathFlag, viper.GetString(tlsKeyPathKey), "the path to the TLS key")
	cmd.PersistentFlags().Bool(noTLSFlag, viper.GetBool(noTLSKey), "if present, do not use TLS")
	cmd.PersistentFlags().String(AuthHeaderFlag, viper.GetString(authHeaderKey), "Auth header in the form 'Bearer <base64>'")
}

// GetConfigCommand :
func GetConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config {set,get,delete,init} [args]",
		Short: "Manage the CLI configuration",
	}
	cmd.AddCommand(getConfigGetCommand())
	cmd.AddCommand(getConfigSetCommand())
	cmd.AddCommand(getConfigDeleteCommand())
	cmd.AddCommand(getConfigInitCommand())
	return cmd
}

func getConfigGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "get <key>",
		Short:     "Get CLI option value",
		Args:      cobra.ExactArgs(1),
		ValidArgs: configOptions,
		RunE:      runConfigGetCommand,
	}
}

func runConfigGetCommand(_ *cobra.Command, args []string) error {
	value := viper.Get(args[0])
	_, _ = fmt.Fprintln(GetOutput(), value)
	return nil
}

func getConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "set <key> <value>",
		Short:     "Set CLI option value",
		Args:      cobra.ExactArgs(2),
		ValidArgs: configOptions,
		RunE:      runConfigSetCommand,
	}
}

func runConfigSetCommand(_ *cobra.Command, args []string) error {
	viper.Set(args[0], args[1])
	if err := viper.WriteConfig(); err != nil {
		return err
	}

	value := viper.Get(args[0])
	_, _ = fmt.Fprintln(GetOutput(), value)
	return nil
}

func getConfigDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "delete <key>",
		Short:     "Delete CLI option value",
		Args:      cobra.ExactArgs(1),
		ValidArgs: configOptions,
		RunE:      runConfigDeleteCommand,
	}
}

func runConfigDeleteCommand(_ *cobra.Command, args []string) error {
	viper.Set(args[0], nil)
	if err := viper.WriteConfig(); err != nil {
		return err
	}

	value := viper.Get(args[0])
	_, _ = fmt.Fprintln(GetOutput(), value)
	return nil
}

func getConfigInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: fmt.Sprintf("Initialize the %s CLI configuration", configName),
		RunE:  runConfigInitCommand,
	}
}

func runConfigInitCommand(_ *cobra.Command, _ []string) error {
	return CreateConfig(true)
}

func getCertPath(cmd *cobra.Command) string {
	certPath, _ := cmd.Flags().GetString(tlsCertPathFlag)
	return certPath
}

// InitConfig :
func InitConfig(configNameInit string) {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	configName = configNameInit
	viper.SetConfigName(configNameInit)
	viper.AddConfigPath(home + "/" + configDir)
	viper.AddConfigPath("/etc/onos")

	_ = viper.ReadInConfig()
}
