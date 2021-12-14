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

export INGRESS_CLASS=${INGRESS_CLASS:-istio.ingress.networking.knative.dev}

function is_ingress_class() {
  [[ "${INGRESS_CLASS}" == *"${1}"* ]]
}

function cluster_setup() {
  header "Building client"
  ${REPO_ROOT_DIR}/hack/build.sh -f || return 1
}

# Copied from knative/serving setup script:
# https://github.com/knative/serving/blob/main/test/e2e-networking-library.sh#L17
function install_istio() {
  if [[ -z "${ISTIO_VERSION:-}" ]]; then
    readonly ISTIO_VERSION="latest"
  fi
  header "Installing Istio ${ISTIO_VERSION}"
  local LATEST_NET_ISTIO_RELEASE_VERSION=$(curl -L --silent "https://api.github.com/repos/knative/net-istio/releases" | \
    jq -r '[.[].tag_name] | sort_by( sub("knative-";"") | sub("v";"") | split(".") | map(tonumber) ) | reverse[0]')
  # And checkout the setup script based on that release
  local NET_ISTIO_DIR=$(mktemp -d)
  (
    cd $NET_ISTIO_DIR \
      && git init \
      && git remote add origin https://github.com/knative-sandbox/net-istio.git \
      && git fetch --depth 1 origin $LATEST_NET_ISTIO_RELEASE_VERSION \
      && git checkout FETCH_HEAD
  )

  if [[ -z "${ISTIO_PROFILE:-}" ]]; then
    readonly ISTIO_PROFILE="istio-ci-no-mesh.yaml"
  fi

  if [[ -n "${CLUSTER_DOMAIN:-}" ]]; then
    sed -ie "s/cluster\.local/${CLUSTER_DOMAIN}/g" ${NET_ISTIO_DIR}/third_party/istio-${ISTIO_VERSION}/${ISTIO_PROFILE}
  fi

  echo ">> Installing Istio"
  echo "Istio version: ${ISTIO_VERSION}"
  echo "Istio profile: ${ISTIO_PROFILE}"
  ${NET_ISTIO_DIR}/third_party/istio-${ISTIO_VERSION}/install-istio.sh ${ISTIO_PROFILE}

}

function knative_setup() {
  if is_ingress_class istio; then
    install_istio
  fi

  local serving_version=${KNATIVE_SERVING_VERSION:-latest}
  header "Installing Knative Serving (${serving_version})"

  if [ "${serving_version}" = "latest" ]; then
    start_latest_knative_serving
  else
    start_release_knative_serving "${serving_version}"
  fi

  if ! is_ingress_class istio; then
    kubectl patch configmap/config-network -n knative-serving \
      --type merge -p '{"data": {"ingress.class":"'${INGRESS_CLASS}'"}}'
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
