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

source $(dirname $0)/../vendor/knative.dev/hack/e2e-tests.sh

function cluster_setup() {
  header "Building client"
  ${REPO_ROOT_DIR}/hack/build.sh -f || return 1
}

function install_istio() {
  # Other versions need be added through import in hack/tools.go first
  local ISTIO_VERSION="stable"

  if [[ -z "${ISTIO_PROFILE:-}" ]]; then
    readonly ISTIO_PROFILE="istio-ci-no-mesh.yaml"
  fi

  echo ">> Installing Istio"
  echo "Istio version: ${ISTIO_VERSION}"
  echo "Istio profile: ${ISTIO_PROFILE}"
  chmod +x ./vendor/knative.dev/net-istio/third_party/istio-${ISTIO_VERSION}/install-istio.sh
  ./vendor/knative.dev/net-istio/third_party/istio-${ISTIO_VERSION}/install-istio.sh ${ISTIO_PROFILE} || return 1
}

function knative_setup() {
  install_istio

  local serving_version=${KNATIVE_SERVING_VERSION:-latest}
  header "Installing Knative Serving (${serving_version})"

  if [ "${serving_version}" = "latest" ]; then
    start_latest_knative_serving
  else
    start_release_knative_serving "${serving_version}"
  fi

  local eventing_version=${KNATIVE_EVENTING_VERSION:-latest}
  header "Installing Knative Eventing (${eventing_version})"

  if [ "${eventing_version}" = "latest" ]; then
    start_latest_knative_eventing
    start_latest_eventing_sugar_controller
  else
    start_release_knative_eventing "${eventing_version}"
    start_release_eventing_sugar_controller "${eventing_version}"
  fi
}

# Create test resources and images
function test_setup() {
  echo ">> Uploading test images..."
  ${REPO_ROOT_DIR}/test/upload-test-images.sh || return 1
}
