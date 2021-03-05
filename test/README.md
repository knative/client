# Test

This directory contains e2e tests and testing docs:

- Unit tests are in the code base alongside the code they test
- e2e tests are in [`test/e2e/`](./e2e)

## Running unit tests

To run all unit tests locally:

```bash
$ hack/build.sh -t
```

_By default `go test` will not run [the e2e tests](#running-e2e-tests),
which need [`-tags=e2e`](#running-end-to-end-tests) to be enabled._

## Running E2E tests

To run [the e2e tests](./e2e) locally, you need to have:

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

### Running E2E tests selectively

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
### E2E tests prow jobs

Two e2e tests prow jobs are run in CI:

1. [pull-knative-client-integration-tests](https://prow.knative.dev/job-history/gs/knative-prow/pr-logs/directory/pull-knative-client-integration-tests): Runs client e2e tests with the nightly build of serving and eventing.
2. [pull-knative-client-integration-tests-latest-release](https://prow.knative.dev/job-history/gs/knative-prow/pr-logs/directory/pull-knative-client-integration-tests-latest-release): Runs client e2e tests with the latest release of serving and eventing. The latest release version can be configured [here](presubmit-integration-tests-latest-release.sh).

## Test images

### Building the test images

The [`upload-test-images.sh`](./upload-test-images.sh) script can be used to
build and push the test images used by the e2e tests. The script
expects your environment to be setup as described in
[DEVELOPMENT.md](https://github.com/knative/serving/blob/master/DEVELOPMENT.md#install-requirements).

To run the script for all end to end test images:

```bash
./test/upload-test-images.sh
```

A docker tag may be passed as an optional parameter. This can be useful on
Minikube in tandem with the `--tag` [flag](#using-a-docker-tag):

`PLATFORM` environment variable is optional. If it is specified, test images
will be built for specific hardware architecture, according to its value
(for instance,`linux/arm64`).

```bash
eval $(minikube docker-env)
./test/upload-test-images.sh any-old-tag
```

### Adding new test images

New test images should be placed in `./test/test_images`.
