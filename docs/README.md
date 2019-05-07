# Kn

Kn is the Knative command line interface (CLI). 

It is designed with the goals of:

1. Following closely the Knative [serving](https://github.com/knative/serving) and [eventing](https://github.com/knative/eventing) APIs
2. Being scriptable to allow users to create different Knative workflows
3. Exposing useful Golang packages to allow integration into other programs or CLIs or plugins
4. Use consistent verbs, nouns, and flags that it exposes for each set (groups or categories) of commands
5. To be easily extended via a plugin mechanism (similar to Kubectl) to allow for experimentations and customization

## Command Families

Most Kn commands typically fall into one of a few categories:

| Type                 | Used For                       | Description                                                       |
|----------------------|--------------------------------|-------------------------------------------------------------------|
| Service Management   | Managing Kn services           | List, create, update, and delete a Kn service                     |
| Revision Management  | Managing Kn service revisions  | List, create, update, and delete the revision(s) of a Kn service  |
| Miscellaneous        | Collection of utility commands | Show version of Kn, help, plugin list, and other useful commands  |

## Service Management

A Knative service is the embodiment of a serverless workload. Generally in the form of a collection of containers running in a group of pods in the underlying Kubernetes cluster. Each Knative service associates with a collection of revisions which represents the evolution of that service.

With the Kn CLI a user can [`list`](cmd/kn_service_list.md), [`create`](cmd/kn_service_create.md), [`delete`](cmd/kn_service_delete.md), and [`update`](cmd/kn_service_update.md) Knative services. The [detail reference](cmd/kn_service.md) of each sub-command under the [`service` command](cmd/kn_service.md) shows the options and flags for this group of commands.

Examples:

```bash
# Create a new service from an image

kn service create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest
```

You are able to also specify the requests and limits of both CPU and memory when creating a service. See [`service create` command](cmd/kn_service_create.md) reference for additional details.

```bash
# List existing services in the 'default' namespace of your cluster

kn service list
```

You can also list services from all namespaces or specific namespace using flags: `--all-namespaces` and `--namespace mynamespace`. See [`service list` command](cmd/kn_service_list.md) reference for additional details.

## Revision Management

A Knative revision is a change to the specification of a service. For instance, when a Knative service is created with environment variable `FOO=bar` a revision is added to the service. When later the environment variable is changed to `baz` or additional variables are added, a new revision is created. [What other changes can create revisions?]

With the [`revision` command group](cmd/kn_revision.md) you can [list](cmd/kn_revision_list.md) and [describe](cmd/kn_revision_describe.md) the current revisions on a service.

Examples:

```bash
# Listing a service's revision

kn revision list --service srvc # CHECK this since current command does not have --service flag
```

## Miscellaneous

This is a grab all category for commands that do not fit into the previous categories. We can divide this into two.

### Utilities

These are commands that provide some useful information to the user.

* The `kn help` command displays a list of the commands with helpful information.
* The [`kn version` command](cmd/kn_version.md) will display the current version of the Kn build including date and Git commit revision.
* The [`kn completion` command](cmd/kn_completion.md) will output a BASH completion script for Kn to allow command completions with tabs.

### Plugins

[Plugins](plugins.md) are an experimental feature to allow users to extend and customize the Kn CLI.

## Common Flags

For every Kn command you can use these optional common additional flags:

* `-h` or `--help` to display specific help for that command
* `--config string` which specifies the Kn config file (default is $HOME/.kn.yaml)
* `--kubeconfig string` which specifies the kubectl config file (default is $HOME/.kube/config)