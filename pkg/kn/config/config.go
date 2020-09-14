// Copyright © 2020 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// bootstrapDefaults are the defaults values to use
type defaultConfig struct {
	configFile          string
	pluginsDir          string
	lookupPluginsInPath bool
}

// Initialize defaults
var bootstrapDefaults = initDefaults()

// config contains the variables for the Kn config
type config struct {
	// configFile is the config file location
	configFile string

	// sinkMappings is a list of sink mapping
	sinkMappings []SinkMapping

	// channelTypeMappings is a list of channel type mapping
	channelTypeMappings []ChannelTypeMapping
}

// ConfigFile returns the config file which is either the default XDG conform
// config file location or the one set with --config
func (c *config) ConfigFile() string {
	if c.configFile != "" {
		return c.configFile
	}
	return bootstrapDefaults.configFile
}

// PluginsDir returns the plugins' directory
func (c *config) PluginsDir() string {
	if viper.IsSet(keyPluginsDirectory) {
		return viper.GetString(keyPluginsDirectory)
	} else if viper.IsSet(legacyKeyPluginsDirectory) {
		// Remove that branch if legacy option is switched off
		return viper.GetString(legacyKeyPluginsDirectory)
	} else {
		return bootstrapDefaults.pluginsDir
	}
}

// LookupPluginsInPath returns true if plugins should be also checked in the pat
func (c *config) LookupPluginsInPath() bool {
	if viper.IsSet(keyPluginsLookupInPath) {
		return viper.GetBool(keyPluginsLookupInPath)
	} else if viper.IsSet(legacyKeyPluginsLookupInPath) {
		// Remove that branch if legacy option is switched off
		return viper.GetBool(legacyKeyPluginsLookupInPath)
	} else {
		// If legacy branch is removed, switch to setting the default to viper
		// See TODO comment below.
		return bootstrapDefaults.lookupPluginsInPath
	}
}

func (c *config) SinkMappings() []SinkMapping {
	return c.sinkMappings
}

func (c *config) ChannelTypeMappings() []ChannelTypeMapping {
	return c.channelTypeMappings
}

// Config used for flag binding
var globalConfig = config{}

// GlobalConfig is the global configuration available for every sub-command
var GlobalConfig Config = &globalConfig

// bootstrapConfig reads in config file and boostrap options if set.
func BootstrapConfig() error {

	// Create a new FlagSet for the bootstrap flags and parse those. This will
	// initialize the config file to use (obtained via GlobalConfig.ConfigFile())
	bootstrapFlagSet := flag.NewFlagSet("kn", flag.ContinueOnError)
	AddBootstrapFlags(bootstrapFlagSet)
	bootstrapFlagSet.ParseErrorsWhitelist = flag.ParseErrorsWhitelist{UnknownFlags: true}
	bootstrapFlagSet.Usage = func() {}
	err := bootstrapFlagSet.Parse(os.Args)
	if err != nil && err != flag.ErrHelp {
		return err
	}

	// Bind flags so that options that have been provided have priority.
	// Important: Always read options via GlobalConfig methods
	err = viper.BindPFlag(keyPluginsDirectory, bootstrapFlagSet.Lookup(flagPluginsDir))
	if err != nil {
		return err
	}
	err = viper.BindPFlag(keyPluginsLookupInPath, bootstrapFlagSet.Lookup(flagPluginsLookupInPath))
	if err != nil {
		return err
	}

	// Check if configfile exists. If not, just return
	configFile := GlobalConfig.ConfigFile()
	_, err = os.Lstat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file to read
			return nil
		}
		return errors.Wrap(err, fmt.Sprintf("cannot stat configfile %s", configFile))
	}

	viper.SetConfigFile(GlobalConfig.ConfigFile())
	viper.AutomaticEnv() // read in environment variables that match

	// Defaults are taken from the parsed flags, which in turn have bootstrap defaults
	// TODO: Re-enable when legacy handling for plugin config has been removed
	// For now default handling is happening directly in the getter of GlobalConfig
	// viper.SetDefault(keyPluginsDirectory, bootstrapDefaults.pluginsDir)
	// viper.SetDefault(keyPluginsLookupInPath, bootstrapDefaults.lookupPluginsInPath)

	// If a config file is found, read it in.
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	// Deserialize sink mappings if configured
	err = parseSinkMappings()
	if err != nil {
		return err
	}

	// Deserialize channel type mappings if configured
	err = parseChannelTypeMappings()
	return err
}

// Add bootstrap flags use in a separate bootstrap proceeds
func AddBootstrapFlags(flags *flag.FlagSet) {
	flags.StringVar(&globalConfig.configFile, "config", "", fmt.Sprintf("kn configuration file (default: %s)", defaultConfigFileForUsageMessage()))
	flags.String(flagPluginsDir, "", "Directory holding kn plugins")
	flags.Bool(flagPluginsLookupInPath, false, "Search kn plugins also in $PATH")

	// Let's try that and mark the flags as hidden: (as those configuration is a permanent choice of operation)
	flags.MarkHidden(flagPluginsLookupInPath)
	flags.MarkHidden(flagPluginsDir)
}

// ===========================================================================================================

// Initialize defaults. This happens lazily go allow to change the
// home directory for e.g. tests
func initDefaults() *defaultConfig {
	return &defaultConfig{
		configFile:          defaultConfigLocation("config.yaml"),
		pluginsDir:          defaultConfigLocation("plugins"),
		lookupPluginsInPath: false,
	}
}

func defaultConfigLocation(subpath string) string {
	var base string
	if runtime.GOOS == "windows" {
		base = defaultConfigDirWindows()
	} else {
		base = defaultConfigDirUnix()
	}
	return filepath.Join(base, subpath)
}

func defaultConfigDirUnix() string {
	home, err := homedir.Dir()
	if err != nil {
		home = "~"
	}

	// Check the deprecated path first and fallback to it, add warning to error message
	if configHome := filepath.Join(home, ".kn"); dirExists(configHome) {
		migrationPath := filepath.Join(home, ".config", "kn")
		fmt.Fprintf(os.Stderr, "WARNING: deprecated kn config directory '%s' detected. Please move your configuration to '%s'\n", configHome, migrationPath)
		return configHome
	}

	// Respect XDG_CONFIG_HOME if set
	if xdgHome := os.Getenv("XDG_CONFIG_HOME"); xdgHome != "" {
		return filepath.Join(xdgHome, "kn")
	}
	// Fallback to XDG default for both Linux and macOS
	// ~/.config/kn
	return filepath.Join(home, ".config", "kn")
}

func defaultConfigDirWindows() string {
	home, err := homedir.Dir()
	if err != nil {
		// Check the deprecated path first and fallback to it, add warning to error message
		if configHome := filepath.Join(home, ".kn"); dirExists(configHome) {
			migrationPath := filepath.Join(os.Getenv("APPDATA"), "kn")
			fmt.Fprintf(os.Stderr, "WARNING: deprecated kn config directory '%s' detected. Please move your configuration to '%s'\n", configHome, migrationPath)
			return configHome
		}
	}

	return filepath.Join(os.Getenv("APPDATA"), "kn")
}

func dirExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

// parse sink mappings and store them in the global configuration
func parseSinkMappings() error {
	// Parse sink configuration
	key := ""
	if viper.IsSet(keySinkMappings) {
		key = keySinkMappings
	}
	if key == "" && viper.IsSet(legacyKeySinkMappings) {
		key = legacyKeySinkMappings
	}

	if key != "" {
		err := viper.UnmarshalKey(key, &globalConfig.sinkMappings)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error while parsing sink mappings in configuration file %s",
				viper.ConfigFileUsed()))
		}
	}
	return nil
}

// parse channel type mappings and store them in the global configuration
func parseChannelTypeMappings() error {
	if viper.IsSet(keyChannelTypeMappings) {
		err := viper.UnmarshalKey(keyChannelTypeMappings, &globalConfig.channelTypeMappings)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error while parsing channel type mappings in configuration file %s",
				viper.ConfigFileUsed()))
		}
	}
	return nil
}

// Prepare the default config file for the usage message
func defaultConfigFileForUsageMessage() string {
	if runtime.GOOS == "windows" {
		return "%APPDATA%\\kn\\config.yaml"
	}
	return "~/.config/kn/config.yaml"
}
