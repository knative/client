# Knative Client

Knative developer experience best practices, reference Knative CLI
implementation, and reference Knative client libraries.

If you are interested in contributing, see the Knative community [contribution guide](https://www.knative.dev/contributing/) and [DEVELOPMENT.md](./DEVELOPMENT.md).

# Docs

Start with the [user's guide](docs/README.md) and from there you can can read about common use cases, get detail docs on each command, and even how to extend the `kn` CLI. Links below for easy access.

* [User's guide](docs/README.md)
* [Basic workflows](docs/workflows.md) (use cases)
* [Generated documentation](docs/cmd/kn.md)
* [Plugins](docs/plugins.md) motivation and guide

**Bash auto completion:**

Run following to enable bash auto completion

```sh
$ source <(kn completion)
```

Use TAB to list available sub-commands

```sh
$ kn <TAB>
completion revision service

$ kn revision <TAB>
describe list
```
