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
	serving_lib "github.com/knative/client/pkg/serving"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
)

type ConfigurationEditFlags struct {
	Image string
	Env   map[string]string
}

func (p *ConfigurationEditFlags) AddFlags(command *cobra.Command) {
	command.Flags().StringVar(&p.Image, "image", "", "Image to run.")
	command.Flags().StringToStringVar(&p.Env, "env", map[string]string{},
		"Environment, comma-separated NAME=value.")
}

func (p *ConfigurationEditFlags) Apply(config *servingv1alpha1.ConfigurationSpec) (err error) {
	err = serving_lib.UpdateEnvVars(config, p.Env)
	if err != nil {
		return err
	}
	err = serving_lib.UpdateImage(config, p.Image)
	if err != nil {
		return err
	}
	return nil
}
