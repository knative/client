# Plugins [WIP]

Plugins are a means to allow experimentation and customization to the Kn experience.

## Motivation

There are two primary motivations to adding plugins to the Kn CLI.

1. Experimentations - the Kn CLI goals are to follow the Knative serving and eventing APIs. This means that commands will be limited to what is offered in these APIs and some additional utility commands. However, what about common commands that a user finds herself invoking by calling multiple Kubectl commands? e.g., finding the pod(s) for a Knative service. Or what about not yet release features of the APIs? A plugin command could solve these very issues.

2. Customizations - while Kn intends to provide a useful and complete CLI for the Knative APIs, customers of Knative and Kn may have additional commands that they want to add to complete or customize the resulting user experience. For instance, installation and updates commands. Or for instance a `doctor` command to help heal the Knative installation. These commands may be specific to the targeted cloud or Kube cluster. A plugin command could solve these idiosyncratic needs.

Other motivation might exist. And when they surface, from our users' actual usage of the plugin in Kn, we will list them here.

## Use-cases

Since we have divided the motivation for plugins into two general categories, let's explore some known use cases for each category.

### Experimentations

1. `doctor` plugin command - to perform som diagnostics on the Knative cluster and perhaps list actions (or perform them automatically) to heal the cluster.
2. `pods` plugin command - to list the pods for a currently running Knative service.
3. `ssh` plugin command - to automatically ssh into a running pod hosting a Knative service.
4. `monitor` plugin command - to aggregate and stream the logs and thus help monitor a running Knative service.
5. `event-flow` plugin command - to combine a series of events and perform some trigger on a Knative service.

### Customizations

1. `install` plugin command - a cloud-specific install/uninstall command. For instance, installing and uninstalling Knative on the IBM Kubernetes Service (IKS). This needs to make use of IKS-specific commands and setup and credentials. The same could be said for other Kubernetes services in the market.
2. `update` plugin command - a command to update the Knative installation.
3. `migrate` plugin command - a command to migrate services from a proprietary or other OSS FaaS platform to Knative.

## Architecture

The current proposal is to follow the Kubectl plugin architecture. The following is a brief summary of this architecture. 

* Any executable on the terminal's PATH that starts with `kn-`, e.g., `/usr/local/bin/kn-doctor`
* All user parameters and flags are passed to the plugin executable as is
* All environment variables are passed to the plugin executable as is
* A warning is displayed to the user if the command tries to shadow another plugin in a different part of the PATH
* A error is displayed to the user if the command tries to shadow an existing Kn command, e.g., a plugin named `kn-service-list`
* For plugins exposing commands with groups and subcommands, then the name needs to also include dashes. So a plugin `kn pods list` would be named `kn-pods-list` 
* For plugins with names that have dashes, then the name needs to also include underscore. So a plugin `kn pods-list` would be named `kn-pods_list` 

Additional details for this architecture can be found in the Kubectl repository.

## Examples

The following are some simple example plugins ready to try.

### pods (wrapping another CLI as a plugin)

TODO

### doctor (new Golang CLI plugin)

TODO

### monitor (bash CLI plugin)

TODO

## Alternative Implementations

There are many ways to add and create plugins in Golang. We have at least considered two different approaches. We discuss each with their pros and cons.

### Hashicorp's go-plugins

Hashicorp has extensive experience using plugins for their CLIs. They have created a common framework to add plugins to CLI called go-plugins. It's main advantages are that this is a:

 1. mature framework, used in many well known CLIs, e.g., `packer`, `terraform`, etc.
 2. dictates a strict protocol between host and plugin. So allows test for compatibility and upgrades
 3. allows passing of complex data structures, e.g., Knative client object, to the plugin
 4. requires explicit install and uninstall of plugins
 5. allows a unified view of the CLI help

While interesting, it also has some drawbacks. Namely:

 1. requires the use of an RPC mechanism between plugin and host CLI
 2. requires a tight interface between plugin and host CLI
 3. requires sharing objects between host and plugin CLI

 Due to some of these drawbacks and the complexity of creating and hosting plugins vs the approach we have and the fact we do not have a strict interface between the host and the plugin, we have opted not to follow this approach.

### No plugins (just kn-* binaries)

In this approach, the Kn CLI does not host plugins but rather we create a convention (compatible with current approach) that plugins are just named `kn-*` and placed on the PATH.

This approach is compatible with the current proposal but lacks the ability to `list` plugins and passing environment variables and other common information between the Kn CLI and the plugin. It also lacks even more structure and is not prevented with the current approach.

## Consequences

In this section we list some of the potential ramifications of adding a plugin mechanism to Kn. Obviously for the negative consequences we list them in order to warn and provide mitigation steps to prevent them.

### Positive

There are many positive consequences to adding a plugin mechanism to Kn. As mentioned in the Motivation section, plugins allow for experimentations and customization. Obviously we would want to have a core set of features in Kn before enabling a plugin mechanism. However, the point is to encourage adoption and perhaps discover the creative contributions from the community.

### Negative

There are, alas, potentially negative consequences to adding a plugin mechanism to Kn. We are listing some here with the intent to provide or implement some mitigations to prevent them.

1. Fragmentation - could happen if plugins are added to a Knative installation that completely re-writes the "language" of Kn. This could be a bad thing if users start associating these new "language" as Kn.
2. Overriding - while the plugin mechanism prevents directly overriding existing core commands, a plugin named `kn-service_create` is allowed and therefore could confuse users.
3. Security - if plugins before so successful that users install them blindly after installing Knative. In general a new security vector is created and it's one more thing that a Knative user need to be mindful of.