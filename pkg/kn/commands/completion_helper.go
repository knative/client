// Copyright © 2021 The Knative Authors
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
	"strings"

	"github.com/spf13/cobra"
)

type completionConfig struct {
	params     *KnParams
	command    *cobra.Command
	args       []string
	toComplete string
	target     string
}

var (
	resourceToFuncMap = map[string]func(config *completionConfig) ([]string, cobra.ShellCompDirective){
		"service": completeService,
	}
)

// ResourceNameCompletionFunc will return a function that will autocomplete the name of
// the resource based on the subcommand
func ResourceNameCompletionFunc(p *KnParams) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		var use string
		if cmd.Parent() != nil {
			use = cmd.Parent().Use
		}
		config := completionConfig{
			p,
			cmd,
			args,
			toComplete,
			getTargetFlagValue(cmd),
		}
		return config.getCompletion(use)
	}
}

func (config *completionConfig) getCompletion(parent string) ([]string, cobra.ShellCompDirective) {
	completionFunc := resourceToFuncMap[parent]
	if completionFunc == nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
	return completionFunc(config)
}

func getTargetFlagValue(cmd *cobra.Command) string {
	flag := cmd.Flag("target")
	if flag == nil {
		return ""
	}
	return flag.Value.String()
}

func completeGitOps(config *completionConfig) (suggestions []string, directive cobra.ShellCompDirective) {
	suggestions = make([]string, 0)
	directive = cobra.ShellCompDirectiveNoFileComp
	if len(config.args) != 0 {
		return
	}
	namespace, err := config.params.GetNamespace(config.command)
	if err != nil {
		return
	}
	client, err := config.params.NewGitopsServingClient(namespace, config.target)
	if err != nil {
		return
	}
	serviceList, err := client.ListServices(config.command.Context())
	if err != nil {
		return
	}
	for _, sug := range serviceList.Items {
		if !strings.HasPrefix(sug.Name, config.toComplete) {
			continue
		}
		suggestions = append(suggestions, sug.Name)
	}
	return
}

func completeService(config *completionConfig) (suggestions []string, directive cobra.ShellCompDirective) {
	if config.target != "" {
		return completeGitOps(config)
	}

	suggestions = make([]string, 0)
	directive = cobra.ShellCompDirectiveNoFileComp
	if len(config.args) != 0 {
		return
	}
	namespace, err := config.params.GetNamespace(config.command)
	if err != nil {
		return
	}
	client, err := config.params.NewServingClient(namespace)
	if err != nil {
		return
	}
	serviceList, err := client.ListServices(config.command.Context())
	if err != nil {
		return
	}
	for _, sug := range serviceList.Items {
		if !strings.HasPrefix(sug.Name, config.toComplete) {
			continue
		}
		suggestions = append(suggestions, sug.Name)
	}
	return
}
