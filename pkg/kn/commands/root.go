// Copyright © 2018 The Knative Authors
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

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var kubeCfgFile string

// rootCmd represents the base command when called without any subcommands
func NewKnCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kn",
		Short: "Knative client.",
		Long: `Manage your Knative building blokcs:

Serving: Manage your services and release new software to them.
Build: Create builds and keep track of their results.
Eventing: Manage event subscriptions and channels. Connect up event sources.`,
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kn.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubeCfgFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")
	rootCmd.AddCommand(NewServiceCommand())
	rootCmd.AddCommand(NewRevisionCommand())
	return rootCmd
}

func Execute() {
	rootCmd := NewKnCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initKubeConfig)

}

func initKubeConfig() {
	if kubeCfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		kubeCfgFile = filepath.Join(home, ".kube", "config")
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kn" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kn")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
