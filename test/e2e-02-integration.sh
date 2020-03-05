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
  eval e2e_test || fail_test

  success
}

integration_test() {
  header "Running tests for Knative Serving $KNATIVE_SERVING_VERSION and Eventing $KNATIVE_EVENTING_VERSION"

  go_test_e2e -timeout=45m ./test/e2e || fail_test
  success
}

smoke_test() {
  header "Running smoke tests"

  kubectl create ns $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  trap "kubectl delete ns $KN_E2E_SMOKE_TESTS_NAMESPACE" EXIT

  sleep 4 # Wait for the namespace to get initialized by kube-controller-manager

  #TODO: deprecated tests remove once --async is gone
  ./kn service create svc1 --no-wait --image gcr.io/knative-samples/helloworld-go -e TARGET=Knative -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service create svc2 --no-wait --image gcr.io/knative-samples/helloworld-go -e TARGET=Knative -$KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service create hello --image gcr.io/knative-samples/helloworld-go -e TARGET=Knative -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service list hello -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service update hello --env TARGET=kn -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn revision list hello -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service list -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service create hello --force --image gcr.io/knative-samples/helloworld-go -e TARGET=Awesome -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service create foo --force --image gcr.io/knative-samples/helloworld-go -e TARGET=foo -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn revision list -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service list -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service describe hello -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service describe svc1 -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn route list -n $KN_E2E_SMOKE_TESTS_NAMESPACE  || fail_test
  ./kn service delete hello -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service delete foo -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn service list -n $KN_E2E_SMOKE_TESTS_NAMESPACE | grep -q svc1 || fail_test
  ./kn service delete svc1 -n $KN_E2E_SMOKE_TESTS_NAMESPACE || fail_test
  ./kn source list-types || fail_test
  success
}

# Fire up
run $@