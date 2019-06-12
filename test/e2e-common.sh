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

DEFAULT_BUILD=latest

# Helper functions.

# Build kn before integration tests, so we fail fast in case of error.
function cluster_setup() {
  header "Building client"
  ${REPO_ROOT_DIR}/hack/build.sh -u || return 1
}

function knative_setup() {
  start_release_knative_serving "$KNATIVE_SERVING_VERSION"
}

# This is the variable saving the link to the YAML file to create and delete knative serving.
KNATIVE_SERVING_RELEASE_YAML=""

# Install the stable release Knative/serving in the current cluster.
# This method is copied from github.com/knative/test-infra/scripts/e2e-tests.sh, since the original var
# $KNATIVE_SERVING_RELEASE is readonly. In order to make it changeable, we rewrite the method with another
# variable $KNATIVE_SERVING_RELEASE_YAML.
function start_release_knative_serving() {
  VERSION=${1:-$DEFAULT_BUILD}
  if [ "$VERSION" = "$DEFAULT_BUILD" ]; then
      KNATIVE_SERVING_RELEASE_YAML="https://storage.googleapis.com/knative-nightly/serving/latest/serving.yaml"
      start_latest_knative_serving
  else
      KNATIVE_SERVING_RELEASE_YAML="https://storage.googleapis.com/knative-nightly/serving/v$VERSION/serving.yaml"
      start_release_knative_serving ${VERSION}
  fi;
}

# Uninstalls Knative Serving from the current cluster.
function knative_teardown() {
  kubectl delete --ignore-not-found=true -f ${KNATIVE_SERVING_RELEASE_YAML} || return 1
}

function run_tests() {
  # Add local dir to have access to built kn
  export PATH=$PATH:${REPO_ROOT_DIR}
  initialize $@
  header "[Knative serving version of $KNATIVE_SERVING_VERSION]: Running tests"
  go_test_e2e ./test/e2e || fail_test
  success
}
