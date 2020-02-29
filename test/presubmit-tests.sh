#!/usr/bin/env bash

# Copyright 2018 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script runs the presubmit tests; it is started by prow for each PR.
# For convenience, it can also be executed manually.
# Running the script without parameters, or with the --all-tests
# flag, causes all tests to be executed, in the right order.
# Use the flags --build-tests, --unit-tests and --integration-tests
# to run a specific set of tests.

# Markdown linting failures don't show up properly in Gubernator resulting
# in a net-negative contributor experience.
# Tracked by https://github.com/knative/test-infra/issues/428

# If you call this script after configuring the environment variables
# $KNATIVE_SERVING_VERSION / $KNATIVE_EVENTING_VERSION with a valid release,
# e.g. 0.6.0, Knative Serving / Eventing of this specified version will be installed
# in the Kubernetes cluster, and all the tests will run against Knative
# Serving / Eventing of this specific version.

export DISABLE_MD_LINTING=1
export PRESUBMIT_TEST_FAIL_FAST=1

export GO111MODULE=on
export KNATIVE_SERVING_VERSION=${KNATIVE_SERVING_VERSION:-latest}
export KNATIVE_EVENTING_VERSION=${KNATIVE_EVENTING_VERSION:-latest}
source $(dirname $0)/../vendor/knative.dev/test-infra/scripts/presubmit-tests.sh

# Run cross platform build to ensure kn compiles for Linux, macOS and Windows
function post_build_tests() {
  local failed=0
  header "Ensuring kn cross platform build"
  ./hack/build.sh -x || failed=1
  if (( failed )); then
    results_banner "Cross platform build" ${failed}
    exit ${failed}
  fi
}

# Run the unit tests with an additional flag '-mod=vendor' to avoid
# downloading the deps in unit tests CI job
function unit_tests() {
  report_go_test -race -mod=vendor ./... || failed=1
}

# We use the default build and integration test runners.
main $@
