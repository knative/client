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
function cluster_setup() {
  header "Building client"
  go build -v -mod=vendor ./cmd/... || return 1
}

function knative_setup() {
  start_latest_knative_serving
}

# Script entry point.

initialize $@

header "Running tests"

./kn service create hello --image http://gcr.io/knative-samples/helloworld-go -e TARGET=Knative || fail_test
sleep 2
./kn service get || fail_test
./kn service update hello --env TARGET=kn || fail_test
./kn revision get || fail_test
./kn service get || fail_test
./kn service describe hello || fail_test
./kn service delete hello || fail_test

success
