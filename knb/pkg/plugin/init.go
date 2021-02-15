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
				return fmt.Errorf("File '%s' already exists\n", outputFile)
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
		"E.g `kn service log` can be achieved with `--cmd service,log`.")

	return registerCmd
}

var pluginTemplate = `
package plugin

import (
    "errors"
	"os"

	{{if .PluginImportPath}}"{{.PluginImportPath	}}"{{else}}//TODO: add plugin import{{end}}

	"knative.dev/client/pkg/kn/plugin"
)

func init() {
	plugin.InternalPlugins = append(plugin.InternalPlugins, &inlinedPlugin{})
}

type inlinedPlugin struct{}

// Name is a plugin's name
func (p *inlinedPlugin) Name() string {
	return "{{.Name}}"
}

// Execute represents the plugin's entrypoint when called through kn
func (p *inlinedPlugin) Execute(args []string) error {
	//TODO: implement plugin command execution
    //cmd := root.NewPluginCommand()
	//oldArgs := os.Args
	//defer (func() {
	//	os.Args = oldArgs
	//})()
	//os.Args = append([]string{"{{.Name}}"}, args...)
    //return cmd.Execute()
	return errors.New("plugin execution is not implemented yet")
}

// Description is displayed in kn's plugin section
func (p *inlinedPlugin) Description() (string, error) {
	{{if .Description}}return "{{.Description}}", nil{{else}}//TODO: add description
    return "", nil{{end}}
}

// CommandParts defines for plugin is executed from kn
func (p *inlinedPlugin) CommandParts() []string {
	return []string{ {{- range $i,$v := .CmdParts}}{{if $i}}, {{end}}"{{.}}"{{end -}} }
}

// Path is empty because its an internal plugins
func (p *inlinedPlugin) Path() string {
	return ""
}`
