# `kn` Plugins

Plugins follow a similar architecture to [kubectl plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) with some small differences. One key difference is that `kn` plugins can either live in your `PATH` or in a chosen and specified directory. [Kn plugins](docs/cmd/kn_plugin.md) shows how to install and create new plugins as well as giving some examples and best practices.

To see what plugins are installed on your machine, you can use the [`plugin`](docs/cmd/kn_plugin.md) command group's [`list`](docs/cmd/kn_plugin_list.md) command.

Plugins provide extended functionality that is not part of the core `kn` command-line distribution.

Please refer to the documentation and examples for more information about how write your own plugins.
