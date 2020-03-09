# `kn` Plugins

Plugins follow a similar architecture to
[kubectl plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)
with some small differences. One key difference is that `kn` plugins can either
live in your `PATH` or in a chosen and specified directory.
[Kn plugins](https://github.com/knative/client/tree/master/docs/cmd/kn_plugin.md)
shows how to install and create new plugins as well as gives some examples and
best practices.

To see what plugins are installed on your machine, you can use the
[`plugin`](https://github.com/knative/client/tree/master/docs/cmd/kn_plugin.md)
command group's
[`list`](https://github.com/knative/client/tree/master/docs/cmd/kn_plugin_list.md)
command.

Plugins provide extended functionality that is not part of the core `kn`
command-line distribution.

Please refer to the documentation and examples for more information on how to
write your own plugins.

- [kn plugin](../cmd/kn_plugin.md) - Plugin command group
