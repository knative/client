/*
Copyright 2021 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

// DistroConfig represents yaml configuration struct
type DistroConfig struct {
	Plugins []Plugin `yaml:"plugins"`
}

// NewDistroGenerateCmd represents plugin distro command
func NewDistroGenerateCmd() *cobra.Command {
	var config string
	var generateCmd = &cobra.Command{
		Use:   "distro",
		Short: "Generate required files to build `kn` with inline plugins.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(config); os.IsNotExist(err) {
				return fmt.Errorf("kn distro configuration file '%s' doesn't exist", config)
			}
			if _, err := os.Stat("cmd/kn/main.go"); os.IsNotExist(err) {
				return fmt.Errorf("cmd/kn/main.go doesn't exist, make sure the command is executed in knative/client root directory")
			}

			fmt.Println("Generating customized kn distro:")
			registerFile := filepath.Join("pkg", "kn", "root", "plugin_register.go")
			if _, err := os.Stat(registerFile); !os.IsNotExist(err) {
				fmt.Println("⚠️  plugin_register.go file already exists, trying to append imports")
			}

			rawConf, err := ioutil.ReadFile(config)
			if err != nil {
				return err
			}
			conf := &DistroConfig{}
			if err := yaml.Unmarshal(rawConf, conf); err != nil {
				return err
			}
			fmt.Println("✔  config file '" + config + "' processed")

			for _, p := range conf.Plugins {
				importPath := p.PluginImportPath
				if importPath == "" {
					importPath = p.Module + "/plugin"
				}
				if err := appendImport(registerFile, importPath); err != nil {
					return err
				}
				_, err := exec.Command("go", "mod", "edit", "-require", p.Module+"@"+p.Version).Output()
				if err != nil {
					return fmt.Errorf("go mod edit -require failed: %w", err)
				}
				fmt.Println("✔  go.mod require updated")

				if len(p.Replace) > 0 {
					for _, r := range p.Replace {
						_, err := exec.Command("go", "mod", "edit", "-replace", r.Module+"="+r.Module+"@"+r.Version).Output()
						if err != nil {
							return fmt.Errorf("go mod edit -replace failed: %w", err)
						}
						fmt.Println("✔  go.mod replace updated")
					}
				}
			}
			if err := exec.Command("gofmt", "-s", "-w", registerFile).Run(); err != nil {
				return fmt.Errorf("gofmt failed: %w", err)
			}
			return nil
		},
	}
	generateCmd.Flags().StringVarP(&config, "config", "c", ".kn.yaml", "Path to `kn.yaml` config file")
	return generateCmd
}

// appendImport adds specified importPath to plugin registration file.
// New file is initialized if it doesn't exist.
// Warning message is displayed if the plugin import is already present.
func appendImport(file, importPath string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		t, err := template.New("register").Parse(registerTemplate)
		if err != nil {
			return err
		}
		fmt.Println("✔  " + importPath + " added to plugin_register.go")
		return t.Execute(f, importPath)
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if strings.Contains(string(content), importPath) {
		fmt.Println("⚠️  " + importPath + " is already present, no changes made")
		return nil
	}
	hook := "// Add #plugins# import here. Don't remove this line, it triggers an automatic replacement."
	content = bytes.Replace(content, []byte(hook), []byte(fmt.Sprintf("%s\n    _ \"%s\"", hook, importPath)), 1)
	fmt.Println("✔  " + importPath + " added to plugin_register.go")
	return ioutil.WriteFile(file, content, 0644)
}

var registerTemplate = `
// Copyright © 2020 The OpenShift Knative Authors
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

package root

import (
    // Add #plugins# import here. Don't remove this line, it triggers an automatic replacement.
    _ "{{.}}"
)

// RegisterInlinePlugins is an empty function which however forces the
// compiler to run all init() methods of the registered imports
func RegisterInlinePlugins() {}`
