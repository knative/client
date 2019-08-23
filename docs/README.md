# kn

`kn` is the Knative command line interface (CLI).

## Getting Started

### Installing `kn`

You can grab the latest nightly binary executable for:
 * [Max OS X](https://storage.cloud.google.com/knative-nightly/client/latest/kn-darwin-amd64)
 * [Linux AMD 64](https://storage.googleapis.com/knative-nightly/client/latest/kn-linux-amd64)
 * [Windows AMD 64](https://storage.googleapis.com/knative-nightly/client/latest/kn-windows-amd64.exe)

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

----------------------------------------------------------

## Commands

* See the [generated documentation](cmd/kn.md) 
* See the documentation on [managing `kn`](operations/management.md)


## Plugins

Kn supports plugins, which allow you to extend the functionality of your `kn` installation with custom commands as well as shared commands that are not part of the core distribution of `kn`. See the [plugins documentation]((plugins/README.md)) for more information.


## More information on `kn`:

* [Workflows](workflows/README.md)
* [Operations](operations/README.md)
* [Traffic Splitting](traffic/README.md)


