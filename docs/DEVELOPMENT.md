# Development

This doc explains how to set up a development environment so you can get started
[contributing](https://www.knative.dev/contributing/) to `Knative Client`. Also
take a look at:

- [The pull request workflow](https://www.knative.dev/contributing/reviewing/)

## Prerequisites

Follow the instructions below to set up your development environment. Once you
meet these requirements, you can make changes and
[build your own version of Knative Client](#building-knative-client)!

Before submitting a PR, see also
[contribution guide](https://www.knative.dev/contributing/).

### Sign up for GitHub

Start by creating [a GitHub account](https://github.com/join), then set up
[GitHub access via SSH](https://help.github.com/articles/connecting-to-github-with-ssh/).

### Install requirements

You must install these tools:

1. [`go`](https://golang.org/doc/install): The language `Knative Client` is
   built in (1.13 or later)
1. [`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports)
1. `gcc` compiler: Used during testing. Not needed if golang is installed via
   the installer
1. [`git`](https://help.github.com/articles/set-up-git/): For source control
1. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/): For
   managing development environments

### Create a cluster

1. [Set up Knative](https://knative.dev/docs/install/any-kubernetes-cluster)

### Checkout your fork

To check out this repository:

1. Create your own
   [fork of this repo](https://help.github.com/articles/fork-a-repo/)
1. Clone it to your machine:

```sh
git clone git@github.com:${YOUR_GITHUB_USERNAME}/client.git
cd client
git remote add upstream git@github.com:knative/client.git
git remote set-url --push upstream no_push
```

_Adding the `upstream` remote sets you up nicely for regularly
[syncing your fork](https://help.github.com/articles/syncing-a-fork/)._

Once you reach this point you are ready to do a full build and test as described
below.

## Building Knative Client

Once you've [set up your development environment](#prerequisites), let's build
`Knative Client`.

**Dependencies:**

[go mod](https://github.com/golang/go/wiki/Modules#quick-start) is used and
required for dependencies.

**Building:**

```sh
$ hack/build.sh
```

You can link that script into a directory within your search `$PATH`. This
allows you to build `kn` from any working directory. There are several options
to support various development flows:

- `build.sh` - Compile, test, generate docs and format source code
- `build.sh -f` - Compile only
- `build.sh -f -t` - Compile & test
- `build.sh -c` - Update dependencies, regenerate documentation and format
  source files
- `build.sh -w` - Enter watch mode for automatic recompilation
- `build.sh -w -t` - Enter watch mode for automatic recompilation & running
  tests

See `build.sh --help` for a full list of options and usage examples.

In the end, the build results in `kn` binary in your current directory, which
can be directly executed.

**Testing:**

Please follow the [guide](../test/README.md) here to test the `knative client`.

**Notes:**

- For building, Go `1.11.4` is required
  [due to a go mod issue](https://github.com/golang/go/issues/27925).
- If you are building in your `$GOPATH` folder, you need to specify
  `GO111MODULE` for building it

```sh
# if you are building in your $GOPATH
GO111MODULE=on go build ./cmd/...
```

You can now try updating code for client and test out the changes by building
the `kn` binary.
