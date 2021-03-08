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
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

// NewPluginInitCmd represents plugin init command
func NewPluginInitCmd() *cobra.Command {
	plugin := &Plugin{}
	var outputDir string

	var registerCmd = &cobra.Command{
		Use:   "init",
		Short: "Generate required resource to inline plugin.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := os.MkdirAll(outputDir, os.ModePerm)
			if err != nil {
				return err
			}
			outputFile := filepath.Join(outputDir, "plugin.go")
			if _, err := os.Stat(outputFile); err == nil {
				return fmt.Errorf("file '%s' already exists", outputFile)
			}
			fmt.Println("Generating plugin inline file:")
			f, err := os.Create(outputFile)
			if err != nil {
				return err
			}
			t, err := template.New("init").Parse(pluginTemplate)
			if err != nil {
				return err
			}
			err = t.Execute(f, plugin)
			if err != nil {
				return err
			}
			fmt.Println("✔  plugin inline file generated " + outputFile)
			err = f.Close()
			return err
		},
	}
	registerCmd.Flags().StringVar(&plugin.Name, "name", "", "Name of a plugin.")
	registerCmd.Flags().StringVar(&plugin.Description, "description", "", "Description of a plugin.")
	registerCmd.Flags().StringVar(&plugin.PluginImportPath, "import", "", "Import path of plugin.")
	registerCmd.Flags().StringVar(&outputDir, "output-dir", "plugin", "Output directory to write plugin.go file.")
	registerCmd.Flags().StringSliceVar(&plugin.CmdParts, "cmd", []string{}, "Defines command parts to execute plugin from kn. "+
		"E.g. `kn service log` can be achieved with `--cmd service,log`.")

	return registerCmd
}
