# kn

`kn` is the Knative command-line interface (CLI).

## Getting Started

### Installing `kn`

You can grab the latest nightly binary executable for:

- [macOS](https://storage.googleapis.com/knative-nightly/client/latest/kn-darwin-amd64)
- [Linux](https://storage.googleapis.com/knative-nightly/client/latest/kn-linux-amd64)
- [Windows](https://storage.googleapis.com/knative-nightly/client/latest/kn-windows-amd64.exe)

Please put it on your system path, and make sure it's executable.

Alternatively, check out the client repository, and type:

```bash
go install ./cmd/kn
```

To use the kn container image:

- Nightly: `gcr.io/knative-nightly/knative.dev/client/cmd/kn`
- Latest release: `gcr.io/knative-releases/knative.dev/client/cmd/kn`

To install `kn` using [Homebrew](https://brew.sh):

```bash
brew tap knative/client
brew install kn
```

### Connecting to your cluster

You'll need a `kubectl`-style config file to connect to your cluster.

- Starting [minikube](https://github.com/kubernetes/minikube) writes this file
  (or gives you an appropriate context in an existing config file)
- Instructions for Google
  [GKE](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl)
- Instructions for Amazon
  [EKS](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html)
- Instructions for IBM
  [IKS](https://cloud.ibm.com/docs/containers?topic=containers-getting-started)
- Instructions for Red Hat
  [OpenShift](https://docs.openshift.com/container-platform/4.1/cli_reference/administrator-cli-commands.html#create-kubeconfig)
- Or contact your cluster administrator

`kn` will pick up your `kubectl` config file in the default location of
`$HOME/.kube/config`. You can specify an alternate kubeconfig connection file
with `--kubeconfig`, or the env var `$KUBECONFIG`, for any command.

## Configuration

There are a set of configuration parameters you can set up to customize
`kn`. For example, you can specify where your `kn` plugins are located and how
they are found or add prefixes for Knative eventing sink mappings.

### Location

You can customize your `kn` CLI setup by creating a `config.yaml` configuration file. You can either provide the configuration via a `--config` flag or is picked up from a default location. The default configuration location conforms to the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) and is different for Unix systems and Windows systems. If `XDG_CONFIG_HOME` env var is set, the default config location `kn` looks for is `$XDG_CONFIG_HOME/kn`, otherwise the file is looked up under the home directory of the user at `$HOME/.config/kn/config.yaml`. For Windows systems, the default `kn` configuration location is `%APPDATA%\kn`.

`kn` does not create a default configuration file, nor does it write into an existing configuration file.

### Options

Options in `config.yaml` are grouped into categories:

* **Global** - Global configuration values that affect all `kn`
* **Plugins** - Plugin related configuration values like where to find plugins (sedction `plugins:`)
* **Eventing** - Configuration that impacts eventing related operations (section: `eventing:`)

The following example contains a fully commented `config.yaml` with all available options:

```yaml
# Plugins related configuration
plugins:
  # Whether to lookup configuration in the execution path (default: false)
  path-lookup: true
  # Directory from where plugins are looked up. (default: "$base_dir/plugins"
  # where "$base_dir" is the directory where this configuration file is stored)
  directory: ~/.kn/plugins
# Eventing related configuration
eventing:
  # List of sink mappings that allow custom prefixes wherever a sink
  # specification is used (like for the --sink option of a broker)
  sink-mappings:
    # Prefix as used in the command (e.g. "--sink svc:myservice")
  - prefix: svc
    # Api group of the mapped resource
    group: core
    # Api version of the mapped resource
    version: v1
    # Resource name (lowercased plural form of the 'kind')
    resource: services
  # Channel mappings that you can use in --channel options
  channel-type-mappings:
    # Alias that can be used as a type for a channel option (e.g. "kn create channel mychannel --type Kafka")
  - alias: Kafka
    # Api group of the mapped resource
    group: messaging.knative.dev
    # Api version of the mapped resource
    version: v1alpha1
    # Kind of the resource
    kind: KafkaChannel
```

#### Plugin configuration

You can specify plugin related configuration in the top-level `plugins:` sections.
You can specify the following options:

* `directory`, which is the same as the persistent flag `--plugins-dir` and specifies the kn plugins directory. It defaults to: `~/.config/kn/plugins`.
   By using the persistent flag (when you issue a command) or by specifying the
   value in the `kn` config, a user can select which directory to find `kn`
   plugins. It can be any directory that is visible to the user.

* `path-lookup`, which is the same as the persistent flag
   `--lookup-plugins-in-path` and specifies if `kn` should look for plugins anywhere in the specified `PATH` environment variable. This option is a boolean type, and the default value is `false`.

#### Eventing configuration

The `eventing:` top-level configuration section holds configuration options that influence the behaviour of eventing related commands.

* `sink-mappings` defines prefixes to refer to Kubernetes _Addressable_ resources as
   used in `--sink` kind of options. `sink-mappings refers to an array of individual mapping definitions that you can configure with these fields:
   - `prefix`: Prefix you want to describe your sink as. `ksvc`
      (`serving.knative.dev/v1`), `broker` (`eventing.knative.dev/v1`) and `channel` (`messaging.knative.dev/v1`) are predefined prefixes. Values in the configuration file can override these predefined prefixes.
      - `group`: The APIGroup of Kubernetes resource.
      - `version`: The version of Kubernetes resources.
      - `resource`: The plural name of Kubernetes resources (for example:
         `services`).

* `channel-type-mappings` can be used to define aliases for custom channel types that can be used wherever a channel type is required (as in `kn channel create --type`). This configuration section defines an array of entries with the following fields:
   - `alias`: The name that can be used as the type
   - `group`: The APIGroup of the channel CRD.
   - `version`: The version of the channel CRD.
   - `kind`: Kind of the channel CRD (e.g. `KafkaChannel`)
   -
## Commands

- See the [generated documentation](cmd/kn.md)
- See the documentation on [managing `kn`](operations/management.md)

## Plugins

Kn supports plugins, which allow you to extend the functionality of your `kn`
installation with custom commands and shared commands that are not part
of the core distribution of `kn`. See the
[plugins documentation](plugins/README.md) for more information.

## More information on `kn`:

- [Workflows](workflows/README.md)
- [Operations](operations/README.md)
- [Traffic Splitting](traffic/README.md)
