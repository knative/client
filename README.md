# Kn



[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/knative.dev/client)
[![Go Report Card](https://goreportcard.com/badge/knative/client)](https://goreportcard.com/report/knative/client)
[![Releases](https://img.shields.io/github/release-pre/knative/client.svg)](https://github.com/knative/client/releases)
[![LICENSE](https://img.shields.io/github/license/knative/client.svg)](https://github.com/knative/client/blob/master/LICENSE)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://knative.slack.com)

The Knative client `kn` is your door to the [Knative](https://knative.dev)
world. It allows you to create Knative resources interactively from the command
line or from within Shell scripts.

`kn` offers you:

- Full support for managing all features of
  [Knative Serving](https://github.com/knative/serving) (services, revisions,
  traffic splits)
- Growing support [Knative eventing](https://github.com/knative/eventing),
  closely following its development (managing of sources & triggers)
- A plugin architecture similar to that of `kubectl` plugins
- A thin client-specific API in golang which helps in tasks like synchronously
  waiting on Knative service write operations.
- An easy integration of Knative into Tekton Pipelines by using
  [`kn` in a Tekton `Task`](https://github.com/tektoncd/catalog/tree/master/kn).

This client uses the
[Knative Serving](https://github.com/knative/docs/blob/master/docs/serving/spec/knative-api-specification-1.0.md)
and
[Knative Eventing](https://github.com/knative/eventing/tree/master/docs/spec)
API exclusively so that it will work with any Knative installation, even those
that are not Kubernetes based. It does not help in _installing_ Knative itself
though. Please refer to the various
[Knative installation options](https://knative.dev/docs/install/) for how to
install Knative with its prerequisites.

## Documentation

Start with the [user's guide](docs/README.md) to learn more. You can read about
common use cases, get detailed documentation on each command, and learn how to
extend the `kn` CLI. For more information, have a look at:

- [User guide](docs/README.md)
  - Installation - How to install `kn` and run on your machine
  - Examples - Use case based examples
  - FAQ (_to come._)
- [Reference Manual](docs/cmd/kn.md) - all possible commands and options with
  usage examples

## Developers

We love contributions! Please refer to
[CONTRIBUTING](https://knative.dev/contributing/) for more information on how to
best contributed to contribute to Knative.

For code contributions it as easy as picking an
[issue](https://github.com/knative/client/issues) (look out for
"kind/good-first-issue"), briefly comment that you would like to work on it,
code, test, code and finally submit a
[PR](https://github.com/knative/client/pulls) which will trigger the review
process.

More details on how to build and test can be found in the
[Developer guide](docs/DEVELOPMENT.md).
