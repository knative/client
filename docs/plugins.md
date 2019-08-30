# TL;DR

Similar to `kubectl` (the client CLI for Kubernetes) Kn now supports plugins. The plugin model for Kn was based on `kubectl` plugins but is different in some specific aspects—especially how plugins are executed and located on the user’s machine. In this document we give an overview of the Kn plugin model, highlight the differences with `kubectl` plugins and give some simple examples of Kn plugins in different languages to help you get started creating your own Kn plugins.

## Introduction

The Kn client aims to provide a command line interface (CLI) that is faithful to the Knative APIs. As such all Kn commands must have corresponding backend objects that they represent and expose to the user. There are a few commands that do not fit this philosophy, e.g., `help` and `completion`, however, these are the exceptions rather than the rule. 

### Motivation

This Kn design point raises various issues and questions such as: How to add esoteric or otherwise useful commands that users need for their workflows? How to add commands that are specific to Knative service providers? How to add meta-commands to install or upgrade and manage the Knative installation?

To address these issues and questions and provide a general mechanism for extension, the Knative client team decided to add support for plugins to the Kn client. This mechanism solves similar problems as for the Kubernetes client (`kubectl`) and as such we have created Kn plugins to follow a similar architecture.

### Model

The `kubectl` plugin model is simple and quite powerful. In a nutshell, any executable placed in the user’s `PATH` whose name starts with prefix `kubectl-` is a plugin and will be invoked as part of `kubectl` itself. So an executable named `kubectl-hello` that is in the user’s `PATH` can be invoked with `kubectl` as:

```bash
kubectl hello 
``` 

When invoking the command above, `kubectl` will invoke `kubectl-hello` and pass any parameters that follow the `kubectl hello [params]` invocation to the plugin executable. If the user’s machine has multiple executable files named `kubectl-hello` in different parts of their system, the invocation fails and will output a message. Similarly, plugins cannot be named in a way that conflicts with the current `kubectl` commands.

For Kn we have adopted a similar model except our prefix for naming plugins is `kn-`. So a Kn plugin that is similar to the one discussed above would be named `kn-hello` and be invoked with `kn hello`. Naming conflicts also occur like for `kubectl`. 

However, we have one main difference with `kubectl` plugins and that is the location of Kn plugins can be centralized to a user’s provided location. The default is `~/.kn/plugins` and can be specified in the Kn config file or via the command line using `—lookup-plugins—in-path`. 

You can also override the location of plugins from the default of `~/.kn/plugins` to a different directory by passing the directory with `--plugins-dir` or setting that value in your Kn config. Read about [Kn’s config](docs/config.md) file for details on where it is located and the values you can change.

Before showing some example Kn plugins in different languages, let’s first explore some use cases of possible Kn plugins.

### Use Cases

There are many use cases for Kn Plugins. We give a brief summary of five classes of use cases and some simple examples below to help you get started:

1. *Workflow* to allow users of Knative to expand their current workflows in a natural fashion. For instance DevOps users that end up creating lots of custom scripts for recurring work items, e.g., setting up clusters with Knative on their private cloud, could easily create a `kn-setup` plugin that would blend into the current Kn CLI for ease of use.

2. *Integration* to allow integration of Knative with other systems. For instance a set of plugins to facilitate an integrated and seamless usage of Knative with the [Tekton CI/CD](https://github.com/tektoncd/pipeline).

3. *Extension* to add new features to Knative. For instance creating a `kn-logs` plugin that will use `kubectl` to discover and stream the logs from the pods of deployed and running Knative services.

4. *Tooling* to provide utilities and tools for Knative. For instance to return and display the healthiness of a running Knative cluster using heuristics and the Kubernetes APIs.

5. *Experimentation* to explore the boundaries of what _serverless_ applications and architectures are able to achieve when using Knative as their base platform. For example, creating a plugin to spin up Knative services across clusters in different geographic regions and load balancing taking into account request headers and geo locations.

Naturally since plugins allow a wide range of extensions to Kn and Knative, from a user’s perspective, the use cases listed above constitute only a cross section of what is possible. As Knative and the Kn CLI gain users, we fully anticipate innovative and unexpected plugins to appear in the community.

## Examples

The following simple examples (in different languages) are designed to give you a working plugin that you can test right now on your Knative installation. You can then modify it to create your own plugins.

Please note that these examples are not prescriptive and that any language that allows you to create an executable on your system will do. Also specific languages might be best suited for some class of plugins but not limited to that.

### BASH

The simplest of plugins can be written in BASH

```bash
➜  client git:(master) cat ~/.kn/plugins/kn-hello
#!/bin/bash

echo "Hello Knative, I'm a Kn plugin"
echo "  My plugin file is $0"
echo "  I recieved argument: $1"

➜  client git:(master) kn hello world
Hello Knative, I am a Kn plugin
  My plugin file is /Users/maximilien/.kn/plugins/kn-hello
  I recieved arguments: world
```

### Python

As scripting languages Python and Ruby are ideal for workflow-type plugins and tooling and experimentation. The following plugins showcase the same plugin in both languages.

The plugin `kn-health` uses `kubectl` to query the current Knative objects on a Knative installation for their health and summarizes the result. The goal is to give some idea of the healthiness of the Knative installation.

A sketch of this plugin in Python follows. Complete [source code ](examples/plugins/python/kn-health.py) for the Python version.

```python

```

### Ruby

This is the same `kn-health` plugin written in Ruby. Complete [source code ](examples/plugins/ruby/kn-health.rb) for the Ruby version.

```ruby

```


### Golang

Since the Kn code base is itself in Golang, plugins written in this language are perhaps best for extensions. The resulting code can be structured like Kn itself which could lead to easier  possible integration into the Kn codebase in the future.

The `kn-logs` tries to discover the pods for a running Knative service and aggregate the recent events and displays them to the user. The sketch of the Golang code follows with [complete source here](examples/plugins/golang/kn-logs).

```golang

```

## Next steps

The obvious next step for Kn plugins is to have users of Kn try and create their own plugins and provide feedback to the community under the `#kn-plugins` slack channel. However, there are more work to be done to make plugins usable for developers and users wanting to take Knative and Kn in production. 

Specifically, we need a way to discover and manage plugins. Both at the level of the user’s local machine but also at a community level. Let’s explore each of these as potential future next steps and also discuss the notion of “server-side plugins”.

### Discovery & Management

You can discover the list of plugins visible in your local system by using the `kn plugin list` command. It will display all plugins visible to the current shell executing `kn` and you can ask it to change the location of where it looks for plugins and display more or less information as it searches for plugin. Please see the [kn plugin list](docs/cmd/kn_plugin_list.md) docs for details.

However, what about discovering useful plugins from the community? Short of searching the internet or GitHub for Kn-related plugins it will be difficult for users of Knative and the Kn CLI to easily discover, share, and install remote plugins. This is not a new topic and the `kubectl` community has had to also address it. 

The Kubernetes community has created the [`krew`](https://github.com/kubernetes-sigs/krew) package manager tool as one way to discover, package, and manage installation of `kubectl` plugins. We will be exploring whether extending `krew` is the best approach for Kn plugins or if we should create our own tool for discovery and managing plugins.

### Server-side Plugins

Another possible direction of future improvement is to add support for what we call “server-side plugins”. These are plugins that would be offered from the Knative service providers to users of Knative clusters that are running on their cloud. 

For instance, if you use a Knative cluster offered by `Vendor A` then that vendor might have a series of plugins that extend the Kn CLI to allow you to better manage the cluster. For instance, allowing repairs, easy growth or shrinking of the cluster, using the same Kn CLI experience.

Since these plugins are specific to `Vendor A`’s cloud environment, they are server-side. These plugins might also add additional objects (e.g., Custom Resource Definitions or CRDs) on the user’s Knative installation in order to achieve their purpose.