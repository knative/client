# kn

`kn` is the Knative command line interface (CLI). 

## Getting Started

### Installing `kn`

You can grab the latest nightly binary executable for:
 * [Max OS X](https://storage.cloud.google.com/knative-nightly/client/latest/kn-darwin-amd64)
 * [Linux AMD 64](https://storage.googleapis.com/knative-nightly/client/latest/kn-linux-amd64)
 * [Windows AMD 64](https://storage.googleapis.com/knative-nightly/client/latest/kn-windows-amd64.exe)

Put it on your system path, and make sure it's executable.

Alternately, check out the client repository, and type:

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
 * Or contact your cluster administrator.

`kn` will pick up your `kubectl` config file in the default location of
`$HOME/.kube/config`. You can specify an alternate kubeconfig connection file
with `--kubeconfig`, or the env var `$KUBECONFIG`, for any command.

## Commands

See the [generated documentation.](cmd/kn.md)

### Service Management

A Knative service is the embodiment of a serverless workload. Generally in the form of a collection of containers running in a group of pods in the underlying Kubernetes cluster. Each Knative service associates with a collection of revisions which represents the evolution of that service.

With the Kn CLI a user can list/[`get`](cmd/kn_service_get.md), [`create`](cmd/kn_service_create.md), [`delete`](cmd/kn_service_delete.md), and [`update`](cmd/kn_service_update.md) Knative services. The [detail reference](cmd/kn_service.md) of each sub-command under the [`service` command](cmd/kn_service.md) shows the options and flags for this group of commands.

Examples:

```bash
# Create a new service from an image

kn service create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest
```

You are able to also specify the requests and limits of both CPU and memory when creating a service. See [`service create` command](cmd/kn_service_create.md) reference for additional details.

```bash
# List existing services in the 'default' namespace of your cluster

kn service get
```

You can also list services from all namespaces or specific namespace using flags: `--all-namespaces` and `--namespace mynamespace`. See [`service get` command](cmd/kn_service_get.md) reference for additional details.

### Revision Management

A Knative revision is a "snapshot" of the specification of a service. For instance, when a Knative service is created with environment variable `FOO=bar` a revision is added to the service. When later the environment variable is changed to `baz` or additional variables are added, a new revision is created. When the image the service is running is changed to a new digest, a new revision is created. 

With the [`revision` command group](cmd/kn_revision.md) you can list/[get](cmd/kn_revision_get.md) and [describe](cmd/kn_revision_describe.md) the current revisions on a service.

Examples:

```bash
# Listing a service's revision

kn revision get --service srvc # CHECK this since current command does not have --service flag
```

### Utilities

These are commands that provide some useful information to the user.

* The `kn help` command displays a list of the commands with helpful information.
* The [`kn version` command](cmd/kn_version.md) will display the current version of the `kn` build including date and Git commit revision.
* The [`kn completion` command](cmd/kn_completion.md) will output a BASH completion script for `kn` to allow command completions with tabs.

### Common Flags

For every Kn command you can use these optional common additional flags:

* `-h` or `--help` to display specific help for that command
* `--config string` which specifies the Kn config file (default is $HOME/.kn.yaml)
* `--kubeconfig string` which specifies the kubectl config file (default is $HOME/.kube/config)
