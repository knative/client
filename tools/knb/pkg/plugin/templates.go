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

const registerTemplate = `
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

package root

import (
    // Add #plugins# import here. Don't remove this line, it triggers an automatic replacement.
    _ "{{.}}"
)

// RegisterInlinePlugins is an empty function which however forces the
// compiler to run all init() methods of the registered imports
func RegisterInlinePlugins() {}`

const pluginTemplate = `
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


