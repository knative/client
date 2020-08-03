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

# If you call this script after configuring the environment variable
# $KNATIVE_SERVING_VERSION / $KNATIVE_EVENTING_VERSION with a valid release,
# e.g. 0.10.0, Knative Serving / Eventing of this specified version will be
# installed in the Kubernetes cluster, and all the tests will run against
# Knative Serving / Eventing of this specific version.

source $(dirname $0)/common.sh

# Add local dir to have access to built kn
export PATH=$PATH:${REPO_ROOT_DIR}


run() {
  # Create cluster
  initialize $@

  # Smoke test
  eval smoke_test || fail_test

  # Integration test
  eval integration_test || fail_test

  success
}

integration_test() {
  header "Running tests for Knative Serving $KNATIVE_SERVING_VERSION and Eventing $KNATIVE_EVENTING_VERSION"

  go_test_e2e -timeout=45m ./test/e2e || fail_test
}

smoke_test() {
  header "Running smoke tests"

  # Test namespace
  ns="kne2esmoketests"

  # Test image
  img=${KN_TEST_IMAGE}
  if [[ -z "${KN_TEST_IMAGE}" ]]; then
    img="gcr.io/knative-samples/helloworld-go"
  fi

  set -x

  kubectl create ns $ns || fail_test
  trap "kubectl delete ns $ns" EXIT

  sleep 4 # Wait for the namespace to get initialized by kube-controller-manager

  ./kn service create hello --image $img -e TARGET=Knative -n $ns || kubectl -n knative-serving logs $(kubectl -n knative-serving get pods -oname | grep networking) || fail_test

  set +x
}

# Fire up
run $@
