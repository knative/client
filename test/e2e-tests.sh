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
  initialize $@ --skip-istio-addon

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

smoke_test_clean_up(){
  kubectl delete ns $ns
  kubectl delete ns $ns1
}

smoke_test() {
  header "Running smoke tests"

  # Test namespace
  ns="kne2esmoketests"
  ns1="kne2esmoketest"

  # Test image
  img=${KO_DOCKER_REPO}/helloworld

  set -x

  kubectl create ns $ns || fail_test
  kubectl create ns $ns1 || fail_test
  trap smoke_test_clean_up EXIT

  sleep 4 # Wait for the namespace to get initialized by kube-controller-manager

  ./kn service create svc1 --no-wait --image $img -e TARGET=Knative -n $ns || fail_test
  ./kn service create svc2 --no-wait --image $img -e TARGET=Knative -n $ns || fail_test
  ./kn service create hello --image $img -e TARGET=Knative -n $ns || fail_test
  ./kn service list hello -n $ns || fail_test
  ./kn service update hello --env TARGET=kn -n $ns || fail_test
  ./kn revision list hello -n $ns || fail_test
  ./kn service list -n $ns || fail_test
  ./kn service create hello --force --image $img -e TARGET=Awesome -n $ns || fail_test
  ./kn service create foo --force --image $img -e TARGET=foo -n $ns || fail_test
  ./kn revision list -n $ns || fail_test
  ./kn service list -n $ns || fail_test
  ./kn service describe hello -n $ns || fail_test
  ./kn service describe svc1 -n $ns || fail_test
  ./kn route list -n $ns  || fail_test
  ./kn service delete hello -n $ns || fail_test
  ./kn service delete foo -n $ns || fail_test
  ./kn service list -n $ns | grep -q svc1 || fail_test
  ./kn service delete svc1 -n $ns || fail_test
  ./kn source list-types || fail_test
  ./kn service create mysvc --image $img -e TARGET=Knative -n $ns1 || fail_test
  ./kn source ping create my-ping -n $ns --schedule "*/2 * * * *" --data '{ value: "hello" }' --sink ksvc:mysvc:$ns1 || fail_test
  [[ "$(./kn source ping describe my-ping -n $ns -o=jsonpath={.spec.sink.ref.namespace})" != "$ns1" ]] && fail_test
  ./kn source ping delete my-ping -n $ns || fail_test
  ./kn service delete mysvc -n $ns1 || fail_test

  set +x
}

# Fire up
run $@
