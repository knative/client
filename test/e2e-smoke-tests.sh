#!/usr/bin/env bash

# Copyright 2019 The Knative Authors
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

# This script runs the end-to-end tests for the kn client.

# If you already have the `KO_DOCKER_REPO` environment variable set and a
# cluster setup and currently selected in your kubeconfig, call the script
# with the `--run-tests` argument and it will use the cluster and run the tests.

# Calling this script without arguments will create a new cluster in
# project $PROJECT_ID, start Knative serving, run the tests and delete
# the cluster.

source $(dirname $0)/../vendor/github.com/knative/test-infra/scripts/e2e-tests.sh

# Helper functions.

# Build kn before integration tests, so we fail fast in case of error.
./hack/build.sh -f

function knative_setup() {
  start_latest_knative_serving
}

# Will create and delete this namespace and use it for smoke tests
export KN_E2E_SMOKE_TESTS_NAMESPACE=kne2esmoketests

# Script entry point.

initialize $@

header "Running smoke tests"

kubectl create ns $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test

./kn service create svc1 --async --image gcr.io/knative-samples/helloworld-go -e TARGET=Knative -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service create hello --image gcr.io/knative-samples/helloworld-go -e TARGET=Knative -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service list hello -n $KN_E2E_SMOKE_TESTS_NAMESPACE -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service update hello --env TARGET=kn -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn revision list hello -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service list -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service create hello --force --image gcr.io/knative-samples/helloworld-go -e TARGET=Awesome -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service create foo --force --image gcr.io/knative-samples/helloworld-go -e TARGET=foo -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn revision list -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service list -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service describe hello -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service delete hello -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
./kn service delete foo -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test

kubectl delete ns $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test

success
