# knb - kn builder

CLI tool to enhance plugin build experience for [Knative Client](https://github.com/knative/client).


### Build
```bash
go build
```

### Install
```bash
go get knative.dev/client/tools/knb
```

### Usage

##### Create custom `kn` distribution

The `knb` can be used to generate enhanced customized `kn` source files with inlined plugins.

Create configuration file `.kn.yaml` in a root directory of `knative/client` that should specify at least `name, module, version` coordinates of the plugin.
Executing `knb plugin distro` command will generate the required go files and add dependency to `go.mod`.

Example of `.kn.yaml`
```yaml
plugins:
  - name: kn-plugin-source-kafka
    module: knative.dev/kn-plugin-source-kafka
    pluginImportPath: knative.dev/kn-plugin-source-kafka/plugin
    version: v0.19.0
    replace:
      - module: golang.org/x/sys
        version: v0.0.0-20200302150141-5c8b2ff67527
```

Required:
* name
* module - go module name to be used for import and in go.mod file
* version - accepted values are git tag or branch name of go module.

Optional:
* pluginImportPath - import path override, default `$module/plugin`
* replace - go module replacement defined by `module,version`.


Execute command
```bash
knb plugin distro
```


Build `kn`
```bash
./hack/build.sh
```

##### Enable plugin inline feature

The `knb` can be used to generate required go files to inline any `kn` plugin.

```bash
knb plugin init --name kn-source-kafka --cmd source,kafka --description "Some plugin"
```


##### List of commands

Plugin level commands

```
Manage kn plugins.

Usage:
  knb plugin [command]

Available Commands:
  distro      Generate required files to build `kn` with inline plugins.
  init        Generate required resource to inline plugin.

Flags:
  -h, --help   help for plugin

Use "knb plugin [command] --help" for more information about a command.
```

```
Generate required files to build `kn` with inline plugins.

Usage:
  knb plugin distro [flags]

Flags:
  -c, --config kn.yaml   Path to kn.yaml config file (default ".kn.yaml")
  -h, --help             help for distro

```

```
Generate required resource to inline plugin.

Usage:
  knb plugin init [flags]

Flags:
      --cmd kn service log   Defines command parts to execute plugin from kn. E.g kn service log can be achieved with `--cmd service,log`.
      --description string   Description of a plugin.
  -h, --help                 help for init
      --import string        Import path of plugin.
      --name string          Name of a plugin.
      --output-dir string    Output directory to write plugin.go file. (default "plugin")
```
