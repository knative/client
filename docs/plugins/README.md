# `kn` Plugins

Plugins follow a similar architecture to
[kubectl plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)
with some small differences. One key difference is that `kn` plugins can either
live in your `PATH` or in a chosen and specified directory.
[Kn plugins](https://github.com/knative/client/tree/main/docs/cmd/kn_plugin.md)
show how to install and create new plugins as well as gives some examples and
best practices.

To see what plugins are installed on your machine, you can use the
[`plugin`](https://github.com/knative/client/tree/main/docs/cmd/kn_plugin.md)
command group's
[`list`](https://github.com/knative/client/tree/main/docs/cmd/kn_plugin_list.md)
command.

Plugins provide extended functionality that is not part of the core `kn`
command-line distribution.

Please refer to the documentation and examples for more information on how to
write your own plugins.

- [kn plugin](../cmd/kn_plugin.md) - Plugin command group


## Plugin Inlining

It is possible to inline plugins that are written in golang.
The following steps are required:

* In your plugin project, create a implementation of the `Plugin` interface and add it to the global `plugin.InternalPlugins` slice in your `init()` method, like in this example:

```go
package plugin

import (
	"knative.dev/client/pkg/plugin"
)

func init() {
	plugin.InternalPlugins = append(plugin.InternalPlugins, &myPlugin{})
}
```

* In your fork of the kn client, add a file `plugin_register.go` to the root package directory which imports your plugin's implementation package:

```go
package root

import (
        _ "github.com/rhuss/myplugin/plugin"
)

// RegisterInlinePlugins is an empty function which however forces the
// compiler to run all init() methods of the registered imports
func RegisterInlinePlugins() {}
```

* Update you `go.mod` file with the new dependency and build your custom distribution of `kn`
