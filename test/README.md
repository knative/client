# Test

This directory contains tests and testing docs for `Kn client`:

- [Unit tests](#running-unit-tests) currently reside in the codebase alongside
  the code they test
- [End-to-end tests](#running-end-to-end-tests) :
  - end-to-end tests in [`/test/e2e`](./e2e)

## Running unit tests

To run all unit tests:

```bash
$ hack/build.sh -f -t
```

_By default `go test` will not run [the e2e tests](#running-end-to-end-tests-locally),
which need [`-tags=e2e`](#running-end-to-end-tests) to be enabled._

## Running all end to end tests locally

To run [the e2e tests](./e2e) , you need to have a 

1. [running knative environment.](./../DEVELOPMENT.md#create-a-cluster)
2. kn binary in the $path. 

Before running the e2e tests please make sure you dont have any namespaces with the name starting with `kne2etests`

```bash
$ ./e2e-tests-local.sh
```

### Running a test case selectively

### Running a single test case

To run one e2e test case, e.g. TestAutoscaleUpDownUp, use
[the `-run` flag with `go test`](https://golang.org/cmd/go/#hdr-Testing_flags):

```bash
go test -v -tags=e2e -count=1 ./e2e -run ^TestBasicWorkflow$
```

### Running tests in short mode

Running tests in short mode excludes some large-scale E2E tests and saves
time/resources required for running the test suite. To run the tests in short
mode, use
[the `-short` flag with `go test`](https://golang.org/cmd/go/#hdr-Testing_flags)

```bash
go test -v -tags=e2e -count=1 -short ./e2e
```

### Presubmit tests

Presubmit tests and subsequents tests require a --gcp-project and cannot be run locally.