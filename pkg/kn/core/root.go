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

package core

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/kn/commands/revision"
	"github.com/knative/client/pkg/kn/commands/service"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// rootCmd represents the base command when called without any subcommands
func NewKnCommand(params ...commands.KnParams) *cobra.Command {
	var p *commands.KnParams
	if len(params) == 0 {
		p = &commands.KnParams{}
	} else if len(params) == 1 {
		p = &params[0]
	} else {
		panic("Too many params objects to NewKnCommand")
	}
	p.Initialize()

	rootCmd := &cobra.Command{
		Use:   "kn",
		Short: "Knative client",
		Long: `Manage your Knative building blocks:

Serving: Manage your services and release new software to them.
Build: Create builds and keep track of their results.
Eventing: Manage event subscriptions and channels. Connect up event sources.`,

		// Disable docs header
		DisableAutoGenTag: true,

		// Affects children as well
		SilenceUsage: true,

		// Prevents Cobra from dealing with errors as we deal with them in main.go
		SilenceErrors: true,
	}
	if p.Output != nil {
		rootCmd.SetOutput(p.Output)
	}
	rootCmd.PersistentFlags().StringVar(&commands.KubeCfgFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")

	rootCmd.AddCommand(service.NewServiceCommand(p))
	rootCmd.AddCommand(revision.NewRevisionCommand(p))
	rootCmd.AddCommand(commands.NewCompletionCommand(p))
	rootCmd.AddCommand(commands.NewVersionCommand(p))

	// For glog parse error.
	flag.CommandLine.Parse([]string{})
	return rootCmd
}

func InitializeConfig() {
	cobra.OnInitialize(initKubeConfig)
}

func initKubeConfig() {
	if commands.KubeCfgFile != "" {
		return
	}
	if kubeEnvConf, ok := os.LookupEnv("KUBECONFIG"); ok {
		commands.KubeCfgFile = kubeEnvConf
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		commands.KubeCfgFile = filepath.Join(home, ".kube", "config")
	}
}
