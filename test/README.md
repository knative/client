# Test

This directory contains e2e tests and testing docs:

- Unit tests are in the code base alongside the code they test
- e2e tests are in [`test/e2e/`](./e2e)

## Running unit tests

To run all unit tests:

```bash
$ hack/build.sh -t
```

_By default `go test` will not run [the e2e tests](#running-e2e-tests-locally),
which need [`-tags=e2e`](#running-end-to-end-tests) to be enabled._

## Running e2e tests locally

To run [the e2e tests](./e2e) , you need to have a

1. [Running knative environment.](../docs/DEVELOPMENT.md#create-a-cluster)
2. `kn` binary in the \$PATH.
3. Please Make sure that you are able to connect to the cluster by following the
   [guide here](./../docs#connecting-to-your-cluster)

Before running the e2e tests please make sure you dont have any namespaces with
the name starting with `kne2etests`

Run all e2e tests:

```bash
$ test/local-e2e-tests.sh
```

### Running e2e tests selectively

To run only serving specific e2e tests locally, use

```bash
E2E_TAGS="serving" test/local-e2e-tests.sh
```

To run only eventing specific e2e tests locally, use

```bash
E2E_TAGS="eventing" test/local-e2e-tests.sh
```

### Running a single test case

To run one e2e test case, e.g. TestBasicWorkflow

```bash
test/local-e2e-tests.sh -run ^TestBasicWorkflow$
```

### Running tests in short mode

Running tests in short mode excludes some large-scale E2E tests and saves
time/resources required for running the test suite. To run the tests in short
mode, use
[the `-short` flag with `go test`](https://golang.org/cmd/go/#hdr-Testing_flags)

```bash
test/local-e2e-tests.sh -short
```
