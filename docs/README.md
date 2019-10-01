# kn

`kn` is the Knative command line interface (CLI).

## Getting Started

### Installing `kn`

You can grab the latest nightly binary executable for:
 * [macOS](https://storage.googleapis.com/knative-nightly/client/latest/kn-darwin-amd64)
 * [Linux](https://storage.googleapis.com/knative-nightly/client/latest/kn-linux-amd64)
 * [Windows](https://storage.googleapis.com/knative-nightly/client/latest/kn-windows-amd64.exe)

Put it on your system path, and make sure it's executable.

Alternatively, check out the client repository, and type:

```bash
go install ./cmd/kn
```

### Connecting to your cluster

You'll need a `kubectl`-style config file to connect to your cluster.
 * Starting [minikube](https://github.com/kubernetes/minikube) writes this file
   (or gives you an appropriate context in an existing config file)
 * Instructions for Google [GKE](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl)
 * Instructions for Amazon [EKS](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html)
 * Instructions for IBM [IKS](https://cloud.ibm.com/docs/containers?topic=containers-getting-started)
 * Instructions for Red Hat [OpenShift](https://docs.openshift.com/container-platform/4.1/cli_reference/administrator-cli-commands.html#create-kubeconfig).
 * Or contact your cluster administrator.

`kn` will pick up your `kubectl` config file in the default location of `$HOME/.kube/config`. You can specify an alternate kubeconfig connection file with `--kubeconfig`, or the env var `$KUBECONFIG`, for any command.

## Kn Config

To better customize kn there is a set of configuration parameters you are able to setup and customize. In particular you can specify where your kn plugins are located and how they are found. The kn configuration file is meant to capture these configuration options. Let's explore this file's location, and the options you are able to change with it.

### Location

The default location kn looks for config is under the home directory of user at `$HOME/.kn/config.yaml`. It is not created for you part of Kn installation. You can create this file elsewhere and use `--config` flag to specify its path.

### Options

There are two options you can specify in the kn config file and they are related to how kn locates plugins.

1. `pluginsDir` which is the same as the persistent flag `--plugins-dir` and specifies the kn plugins directory. It defaults to: `~/.kn/plugins`. By using the persistent flag (when you issue a command) or by specifying the value in the kn config, a user can select which directory to find kn plugins. It can be any directory that is visible to the user.

2. `lookupPluginsInPath` which is the same as the persistent flag `--lookup-plugins-in-path` and specficies if kn should look for plugins anywhere in the specified `PATH` environment variable.This is a boolean configuration option and the default value is `false`.

For example, the following kn config will look for kn plugins in the user's `PATH` and also execute plugin in `~/.kn/plugins`.

```bash
cat ~/.kn/config.yaml
lookupPluginsInPath: true
pluginsdir: ~/.kn/plugins
```
----------------------------------------------------------

## Commands

* See the [generated documentation](cmd/kn.md) 
* See the documentation on [managing `kn`](operations/management.md)


## Plugins

Kn supports plugins, which allow you to extend the functionality of your `kn` installation with custom commands as well as shared commands that are not part of the core distribution of `kn`. See the [plugins documentation](plugins/README.md) for more information.


## More information on `kn`:

* [Workflows](workflows/README.md)
* [Operations](operations/README.md)
* [Traffic Splitting](traffic/README.md)

