# Knative Client

Knative developer experience best practices, reference Knative CLI
implementation, and reference Knative client libraries.

If you are interested in contributing, see the Knative community [contribution
guide](https://github.com/knative/docs/blob/master/community/CONTRIBUTING.md).


## How to build

**Dependencies:**

[go mod](https://github.com/golang/go/wiki/Modules#quick-start) is used and required for dependencies

**Requirements:**

  - Golang `1.11.4`

**Building:**

```sh
$ go build ./cmd/...
```

**Notes:**

- knative CLI must be built outside of the $GOPATH folder unless you explicitly use `export GO111MODULE=on`.
- For building, Go `1.11.4` is required [due to a go mod issue](https://github.com/golang/go/issues/27925)


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
