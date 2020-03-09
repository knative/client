# Management Commands

- [Service Management](#Service-Management)
- [Revision Management](#Revision-Management)
- [Utilities](#Utilities)
- [Common Flags](#Common-Flags)

## Service Management

A Knative service is the embodiment of a serverless workload. It is generally in
the form of a collection of containers running in a group of pods, in the
underlying Kubernetes cluster. Each Knative service associates with a collection
of revisions, which represent the evolution of that service.

With the Kn CLI a user can [`list`](../cmd/kn_service_list.md),
[`create`](../cmd/kn_service_create.md),
[`delete`](../cmd/kn_service_delete.md), and
[`update`](../cmd/kn_service_update.md) Knative services. The
[detail reference](../cmd/kn_service.md) of each sub-command under the
[`service`](../cmd/kn_service.md) command shows the options and flags for this
group of commands.

Examples:

```bash
# Create a new service from an image

kn service create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest
```

You are able to also specify the requests and limits of both CPU and memory when
creating a service. See [`service create`](../cmd/kn_service_create.md) command
reference for additional details.

```bash
# List existing services in the 'default' namespace of your cluster

kn service list
```

You can also list services from all namespaces or a specific namespace using
flags: `--all-namespaces` and `--namespace mynamespace`. See
[`service list`](../cmd/kn_service_list.md) command reference for additional
details.

## Revision Management

A Knative revision is a "snapshot" of the specification of a service. For
instance, when a Knative service is created with the environment variable
`FOO=bar` a revision is added to the service. Afterwards, when the environment
variable is changed to `baz` or additional variables are added, a new revision
is created. When the image that the service is running is changed to a new
digest, a new revision is created.

With the [`revision`](../cmd/kn_revision.md) command group you can
[list](../cmd/kn_revision_list.md) and
[describe](../cmd/kn_revision_describe.md) the current revisions on a service.

Examples:

```bash
# Listing a service's revision

kn revision list --service srvc
```

## Utilities

These are commands that provide some useful information to the user.

- The `kn help` command displays a list of the commands with helpful
  information.
- The [`kn version`](../cmd/kn_version.md) command will display the current
  version of the `kn` build including date and Git commit revision.
- The `kn completion` command will output a BASH completion script for `kn` to
  allow command completions with tabs.

## Common Flags

For every `kn` command, you can use these common additional flags:

- `-h` or `--help` to display specific help for that command
- `--config string` which specifies the `kn` config file (default is
  \$HOME/.kn.yaml)
- `--kubeconfig string` which specifies the kubectl config file (default is
  \$HOME/.kube/config)
