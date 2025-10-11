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

1. [Running knative environment.](../DEVELOPMENT.md#create-a-cluster)
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

### Running tests with HTTPS

When running tests against a cluster configured to use HTTPS, use the `--https` flag to make tests expect HTTPS URLs:

```bash
test/local-e2e-tests.sh --https
```



### Running E2E tests as a project admin

It is possible to run the E2E tests by a user with reduced privileges, e.g. project admin.
Some tests require cluster-admin privileges and those tests are excluded from execution in this case.
Running the E2E tests then consists of these steps:
1. The cluster admin creates test namespaces. Each test requires a separate namespace.
By default, the namespace names consist of `kne2etests` prefix and numeric suffix starting from 0:
 `kne2etests0`, `kne2etests1`, etc. The prefix can be configured using the KN_E2E_NAMESPACE env
  variable. The namespace can be created as follows:
    ```bash
    for i in $(seq 0 40); do kubectl create ns "${KN_E2E_NAMESPACE}${i}"; done
    ```
   Note: There are currently slightly over 30 tests but the number will grow so the number of created
   namespaces need to be adjusted.
1. The project admin runs the test suite with specific flags:
    ```bash
    E2E_TAGS="project_admin" test/local-e2e-tests.sh --reusenamespace
    ```
   It is expected that the current user is a project admin for all test namespaces
   and their KUBECONFIG is located at `$HOME/.kube/config` or the env
   variable `$KUBECONFIG` points to it.

### E2E tests prow jobs

Two e2e tests prow jobs are run in CI:

1. [pull-knative-client-integration-tests](https://prow.knative.dev/job-history/gs/knative-prow/pr-logs/directory/pull-knative-client-integration-tests): Runs client e2e tests with the nightly build of serving and eventing.
2. [pull-knative-client-integration-tests-latest-release](https://prow.knative.dev/job-history/gs/knative-prow/pr-logs/directory/pull-knative-client-integration-tests-latest-release): Runs client e2e tests with the latest release of serving and eventing. The latest release version can be configured [here](presubmit-integration-tests-latest-release.sh).

## Test images

### Building the test images

The [`upload-test-images.sh`](./upload-test-images.sh) script can be used to
build and push the test images used by the e2e tests. The script
expects your environment to be setup as described in
[DEVELOPMENT.md](https://github.com/knative/serving/blob/main/DEVELOPMENT.md#install-requirements).

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
