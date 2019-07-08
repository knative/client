# Knative Client

This section outlines best practices for the Knative developer experience, is a reference for Knative CLI
implementation, and a reference for Knative client libraries.

The goals of the Knative Client are to:

1. Follow the Knative [serving](https://github.com/knative/serving) and [eventing](https://github.com/knative/eventing) APIs
2. Be scriptable to allow users to create different Knative workflows
3. Expose useful Golang packages to allow integration into other programs or CLIs or plugins
4. Use consistent verbs, nouns, and flags for various commands
5. Be easily extended via a plugin mechanism (similar to `kubectl`) to allow for experimentation and customization

# Docs

Start with the [user's guide](docs/README.md) to learn more. You can read about common use cases, get detailed documentation on each command, and learn how to extend the `kn` CLI. For more information, access the following links:

* [User's guide](docs/README.md)
* [Basic workflows](docs/workflows.md) (use cases)
* [Generated documentation](docs/cmd/kn.md)

**Bash auto completion:**

Run the following command to enable BASH auto-completion:

```sh
$ source <(kn completion)
```

Use TAB to list available sub-commands:

```sh
$ kn <TAB>
completion revision service version

$ kn revision <TAB>
describe get
```

# Developers

If you would like to contribute, please see
[CONTRIBUTING](https://knative.dev/contributing/)
for more information.

To build `kn`, see our [Development](DEVELOPMENT.md) guide.
