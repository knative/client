# `kn` Developer Guide

This guide described the high-level architectural concepts of the Knative CLI client `kn`.
It is for all developers who contribute to `kn` and want to understand some essential concepts.
For any contributor to `kn`, it is highly recommended reading this document.

_This guide is not complete and has still many gaps that we are going to fill over timer. Contributions are highly appreciated ;-)_

## Flow

The journey starts at [main.go](https://github.com/knative/client/blob/main/cmd/kn/main.go) which is the entry point when you call `kn`.
You find here the main control flow, which can be roughly divided into three phases:

  - _Bootstrap_ is about retrieving essential configuration parameters from command-line flags, and a configuration file.
  - _Plugin Lookup_ finds out whether the user wants to execute a plugin or a built-in command
  - _Execution_ : Execute a plugin or the `RootCommand` for built-in commands

Only the code in the `main` function can use `os.Exit` for leaving the process.
It evaluates any error that bubbles up and prints it out as an error message before exiting.
There is no exception to this rule.

Let's talk now about the three phases separately.

### Bootstrap

The bootstrap performed by [config.BootstrapConfig()](https://github.com/knative/client/blob/0063a263d121702432c1ee71cef30df375a40e76/pkg/config/config.go#L94) extracts all the options relevant for config file detection and plugin configuration.
The bootstrap process does not fully parse all arguments but only those that are relevant for starting up and for looking up any plugin.
The configuration can be either provided via a `--config` flag or is picked up from a default location.
The default configuration location conforms to the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) and is different for Unix systems and Windows systems.

[Viper](https://github.com/spf13/viper) loads the configuration file (if any) form this location, and the `config` package makes all the bootstrap and other configuration information available vial the `config.GlobalConfig` singleton (which btw can be easily mocked for unit tests by just replacing it with another implementation of `config.Config` interface)

### Plugin Lookup

In the next step, a `PluginManager` checks whether the given command-line arguments are pointing to a plugin.
All non-flag arguments are extracted and then used to lookup via [plugin.PluginManager.FindPlugin()](https://github.com/knative/client/blob/0063a263d121702432c1ee71cef30df375a40e76/pkg/plugin/manager.go#L94) in the plugin directory (and the execution `$PATH` if configured) calculated in the _Bootstrap_ phase.

### Execution

If `kn` detects a plugin, it is first validated and then executed.
If not, the flow executes the `RootCommand` which carries all builtin `kn` commands.

However, before `kn` executes a plugin, it is first validated whether it does not clash with a builtin command or command group.
If the validation succeeds, the `PluginManager` calls out to the identified plugin via the `Plugin.Execute()` interface method.

If `kn` does not detect a plugin, the `RootCommand` is executed, which is internally dispatched, based on the given commands, to the matching sub-command.
In the case that no sub-command matches the provided command arguments on the command line, kn throws an error.

## Configuration

For any command that is executed by `kn`, the configuration can come from two places:

* Arguments and flags as provided on the command line
* Configuration stored in a configuration file

`kn` has a _bootstrap configuration_ which consists of these variables:

* `configFile` is the location of the configuration file which by default is `~/.config/kn/config.yaml` on Unix like systems and `%APPDATA%\kn` on Windows of now overridden with the environment variable `XDG_CONFIG_HOME`. It can be specified on the command line with `--config`.
* `pluginsDir` is the location of the directory holding `kn` plugins. By default, this is `${configFile}/plugins`. The default can be overridden in with the field `plugins.directory` in the configuration file.
* `lookupPluginsInPath` is a boolean flag that, when set to `true`, specifies that plugins should also be looked up in the execution `$PATH` (default: `false`). Use `plugins.lookup-path` in the configuration file to override it.


The main flow calls out to `config.BootstrapConfig()` to set these three variables by evaluating the location of the configuration file first and then reading in the configuration.

All other configuration that is stored in the configuration is also available after the bootstrap.
To access this information, any other code can access the methods of `config.GlobalConfig` which holds an implementation of the `config.Config` interface.
For testing purposes, `config.GlobalConfig` can be replaced by a mock implementation while running a test. Use `config.TestConfig` for simple use cases.

## Flags

_This section will hold some information about how flag parsing of commands is supposed to work, including using helper functions for enforcing the conventions defined in [cli conventions](https://github.com/knative/client/blob/main/conventions/cli.md)_

## Clients for accessing the backend

_In this section we are going to describe how commands can access the backend via the `kn` public API which provides simplified access mechanism to the Knative serving and eventing backends_

## Main Packages

Here is a rough overview (as of 2020-06) of the directories hierarchy and packages that organized the `kn` codebase.

### `cmd/kn`

This directory holds the `main` package and is the entry point for calling `kn`.
The `main` package is responsible for the overall flow as described above in the section "Flow".

Potential future CLI commands should be added on the `cmd/` level.

### `pkg/kn`

The directory `kn` holds the code that is specific to the `kn` builtin commands.
Everything below `pkg/kn` is not supposed to be shared outside of `knative/client`.

The following essential packages/directories are stored here:

  - `root` is the package that defines the `RootCommand` which is the top-level [cobra](https://github.com/spf13/cobra) command. The `RootCommand` contains all other command groups and leaf commands.
  - `commands` is a directory that contains all `kn` command definitions that are registered at the `RootCommand`. Here, we distinguish two kinds of commands:
    + **Command Groups** have sub-commands but no code to run on its own. Example: `kn service`
    + **Leaf Commands** have no sub-commands and execute the business logic. Example: `kn service create`
  - `config` is the package for all global configuration handling. The global configuration allows access to the kn configuration file and all top-level options that apply to all commands.
  - `flags` is the package for all flag-parsing utility methods shared across all commands. Some of its content might end up in the `kn public API` packages (see below for more about the `kn` public API).
  - `plugin` has the code for finding, listing and executing plugins.
  - `traffic` has functions for dealing with traffic split.

### `pkg` (APIs)

Besides the `pkg/kn` packages, `pkg` contains packages that are the current public API of `kn` for reuse outside the `knative/client` project.
In the future, these packages are grouped more explicitly in a "public API" package (see below for more on information on this).

The following packages provide simplified access to the corresponding backend API and add some extra functionality like synchronous wait for the completion of an operation like creating a Knative service.

  - `serving` interface to Knative Serving with support for synchronous operations
  - `eventing` interface for access Knative Eventing backend objects like Triggers
  - `dynamic` is a dynamic client for interacting with the backend in a typeless way
  - `sources` has the client access code to the built-in source `ApiServerSource`, `PingSource` and `SinkBinding`.

These packages also determine the API version used for the backend APIs which typically over multiple API versions.
All packages also contain a record-replay style mocking library to be used in Mock testing.

### `pkg` (Misc)

Finally, `pkg` contains also some miscellaneous package for different purposes.
Some of them might end up in the public API, too.

  - `util` : Kitchen-sink for everything that does not find in any other category. Please try to avoid adding files here except for good reasons.
  - `wait` : Helper methods for implementing synchronous calls to backends
  - `printers` : Helper for printing out CLI output in a kn specific way

## Plugins

You can find the central plugin handling in the package `pkg/plugin`.
The entry point here is the `PluginManager`, which is responsible for finding plugins by name and for listing all visible plugins.
An interface `Plugin` describes a plugin.
This interface has currently a single implementation that uses an `exec` system call to run an external plugin command following a naming convention.
Only the `main()` flow looks up and evaluates plugins, except the commands from the `plugin` group.

## Public API

_Describe here the public API package that we also expose to external clients, like e.g. plugins. This public API package is not available yet but will contain `kn` abstraction for accessing backend APIs (serving, eventing, sources, dynamic), the flag-parsing library for enforcing the CLI conventions and any other (utility) code to share with other parties_

## Testing

### Mock Testing

_Describe here the way how our mock testing framework works and that is the preferred way of testing and newly written tests are supposed to use the mock testing style_

### Fake Testing

Fake testing uses the auto-generated fake services provided by the client libraries. Some older tests use these directly but also tests involving the dynamic client. This test couples the code tightly to the external client's API and its version, which we try to avoid so that all our code goes through the client-side APIs that we provide.

_Explain Fake testing more in this section_


