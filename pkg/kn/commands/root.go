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
	"io"
	"os"

	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	cfgFile         string
	rootCmd         *cobra.Command
	config          clientcmd.ClientConfig
	kubeConfigFlags *genericclioptions.ConfigFlags
)

// Parameters for creating commands. Useful for inserting mocks for testing.
type KnParams struct {
	Output         io.Writer
	ServingFactory func() (serving.ServingV1alpha1Interface, error)
}

func (c *KnParams) Initialize() {
	if c.ServingFactory == nil {
		c.ServingFactory = ServingConfig
	}
}

// rootCmd represents the base command when called without any subcommands
func NewKnCommand(params ...KnParams) *cobra.Command {
	var p *KnParams
	if len(params) == 0 {
		p = &KnParams{}
	} else if len(params) == 1 {
		p = &params[0]
	} else {
		panic("Too many params objects to NewKnCommand")
	}
	p.Initialize()

	rootCmd = &cobra.Command{
		Use:   "kn",
		Short: "Knative client",
		Long: `Manage your Knative building blocks:

Serving: Manage your services and release new software to them.
Build: Create builds and keep track of their results.
Eventing: Manage event subscriptions and channels. Connect up event sources.`,
	}
	if p.Output != nil {
		rootCmd.SetOutput(p.Output)
	}

	flags := rootCmd.PersistentFlags()
	kubeConfigFlags = genericclioptions.NewConfigFlags()
	kubeConfigFlags.AddFlags(flags)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kn.yaml)")
	rootCmd.AddCommand(NewServiceCommand(p))
	rootCmd.AddCommand(NewRevisionCommand(p))
	return rootCmd
}

func InitializeConfig() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initKubeConfig)
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
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kn" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kn")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initKubeConfig() {
	config = kubeConfigFlags.ToRawKubeConfigLoader()
	ns, _, err := config.Namespace()
	if err != nil {
		ns = "default"
	}
	rootCmd.PersistentFlags().Set("namespace", ns)
}

func ServingConfig() (serving.ServingV1alpha1Interface, error) {
	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}
	return serving.NewForConfig(clientConfig)
}
