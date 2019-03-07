# Development

This doc explains how to setup a development environment so you can get started
[contributing](https://github.com/knative/docs/blob/master/community/CONTRIBUTING.md)
to `Knative Client`. Also take a look at:

- [The pull request workflow](https://github.com/knative/docs/blob/master/community/CONTRIBUTING.md#pull-requests)

## Prerequisites

Follow the instructions below to set up your development environment. Once you
meet these requirements, you can make changes and
[build your own version of Knative Client](#building-knative-client)!

Before submitting a PR, see also [CONTRIBUTING.md](./CONTRIBUTING.md).

### Sign up for GitHub

Start by creating [a GitHub account](https://github.com/join), then setup
[GitHub access via SSH](https://help.github.com/articles/connecting-to-github-with-ssh/).

### Install requirements

You must install these tools:

1. [`go`](https://golang.org/doc/install): The language `Knative Client` is
   built in (1.11.4 or later)
1. [`git`](https://help.github.com/articles/set-up-git/): For source control
1. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/): For
   managing development environments

### Create a cluster

1. [Set up a Knative](https://github.com/knative/docs/blob/master/install/README.md#install-guides)
   - You can also setup using limited install guides for eg: Minikube/Minishift.

### Setup your environment

To start your environment you'll need to set these environment variables (we
recommend adding them to your `.bashrc`):

1. `GOPATH`: If you don't have one, simply pick a directory and add
   `export GOPATH=...`
1. `$GOPATH/bin` on `PATH`: This is so that tooling installed via `go get` will
   work properly.

`.bashrc` example:

```sh
export GOPATH="$HOME/go"
export PATH="${PATH}:${GOPATH}/bin"
```

### Checkout your fork

The Go tools require that you clone the repository to the
`src/github.com/knative/client` directory in your
[`GOPATH`](https://github.com/golang/go/wiki/SettingGOPATH).

To check out this repository:

1. Create your own
   [fork of this repo](https://help.github.com/articles/fork-a-repo/)
1. Clone it to your machine:

```sh
mkdir -p ${GOPATH}/src/github.com/knative
cd ${GOPATH}/src/github.com/knative
git clone git@github.com:${YOUR_GITHUB_USERNAME}/client.git
cd client
git remote add upstream git@github.com:knative/client.git
git remote set-url --push upstream no_push
```

_Adding the `upstream` remote sets you up nicely for regularly
[syncing your fork](https://help.github.com/articles/syncing-a-fork/)._

Once you reach this point you are ready to do a full build and test as
described below.

## Building Knative Client

Once you've [setup your development environment](#prerequisites), let's build
`Knative Client`.

**Dependencies:**

[go mod](https://github.com/golang/go/wiki/Modules#quick-start) is used and required for dependencies.

**Building:**

```sh
$ GO111MODULE=on go build ./cmd/...
```

It builds `kn` binary in your current directory. You can start playing with it.

**Notes:**

- For building, Go `1.11.4` is required [due to a go mod issue](https://github.com/golang/go/issues/27925).
- If you are building outside of your `$GOPATH` folder, no need to specify `GO111MODULE` for building it

```sh
# if you are building outside your $GOPATH
go build ./cmd/...
```

You can now try updating code for client and test out the changes by building the `kn` binary.
