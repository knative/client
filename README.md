# Knative Client

Knative developer experience best practices, reference Knative CLI
implementation, and reference Knative client libraries.

Goals:

1. Follow closely the Knative [serving](https://github.com/knative/serving) and [eventing](https://github.com/knative/eventing) APIs
2. Be scriptable to allow users to create different Knative workflows
3. Expose useful Golang packages to allow integration into other programs or CLIs or plugins
4. Use consistent verbs, nouns, and flags for various commands
5. Be easily extended via a plugin mechanism (similar to Kubectl) to allow for experimentations and customization

# Docs

Start with the [user's guide](docs/README.md) and from there you can can read about common use cases, get detail docs on each command, and even how to extend the `kn` CLI. Links below for easy access.

* [User's guide](docs/README.md)
* [Basic workflows](docs/workflows.md) (use cases)
* [Generated documentation](docs/cmd/kn.md)

**Bash auto completion:**

Run following to enable bash auto completion

```sh
$ source <(kn completion)
```

Use TAB to list available sub-commands

```sh
$ kn <TAB>
completion revision service version

$ kn revision <TAB>
describe get
```

# Developers

If you'd like to contribute, please see our
[contribution guidelines](https://knative.dev/contributing/)
for more information.

To build `kn`, see our [Development](DEVELOPMENT.md) guide.
