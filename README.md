# Kn

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/knative.dev/client)
[![Go Report Card](https://goreportcard.com/badge/knative/client)](https://goreportcard.com/report/knative/client)
[![Releases](https://img.shields.io/github/release-pre/knative/client.svg)](https://github.com/knative/client/releases)
[![LICENSE](https://img.shields.io/github/license/knative/client.svg)](https://github.com/knative/client/blob/main/LICENSE)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://knative.slack.com)
[![codecov](https://codecov.io/gh/knative/client/branch/main/graph/badge.svg)](https://codecov.io/gh/knative/client)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fknative%2Fclient.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fknative%2Fclient?ref=badge_shield)

The Knative client `kn` is your door to the [Knative](https://knative.dev)
world. It allows you to create Knative resources interactively from the command
line or from within scripts.

`kn` offers you:

- Full support for managing all features of
  [Knative Serving](https://github.com/knative/serving) (services, revisions,
  traffic splits)
- Growing support for [Knative eventing](https://github.com/knative/eventing),
  closely following its development (managing of sources & triggers)
- A plugin architecture similar to that of `kubectl` plugins
- A thin client-specific API in golang which helps with tasks like synchronously
  waiting on Knative service write operations.
- An easy integration of Knative into Tekton Pipelines by using
  [`kn` in a Tekton `Task`](https://github.com/tektoncd/catalog/tree/master/task/kn).


This client uses the
[Knative Serving](https://github.com/knative/docs/blob/main/docs/serving/spec/knative-api-specification-1.0.md)
and
[Knative Eventing](https://github.com/knative/eventing/tree/main/docs/spec)
APIs exclusively so that it will work with any Knative installation, even those
that are not Kubernetes based. It does not help with _installing_ Knative itself
though. Please refer to the various
[Knative installation options](https://knative.dev/docs/install/) for how to
install Knative with its prerequisites.

## Documentation

Refer to the [user's guide](https://knative.dev/docs/client/) to learn more. You can read about
common use cases, get detailed documentation on each command, and learn how to
extend the `kn` CLI.

Following are some useful resources for getting-started using `kn` CLI:

- [Installation](https://knative.dev/docs/client/install-kn/) - how to install `kn` and run on your machine
- [Configuration](https://knative.dev/docs/client/configure-kn/) - how to customize `kn`
- [Reference Manual](docs/cmd/kn.md) - all possible commands and options with
  usage examples

Additionally you can visit [knative.dev](https://knative.dev) for more examples.

## Developers

If you are interested in contributing, see [CONTRIBUTING.md](./CONTRIBUTING.md)
and [DEVELOPMENT.md](DEVELOPMENT.md). For a list of help wanted issues in Knative,
check out [CLOTRIBUTOR](https://clotributor.dev/search?project=knative&page=1)


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fknative%2Fclient.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fknative%2Fclient?ref=badge_large)