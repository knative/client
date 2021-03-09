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
  header "Installing Istio (${KNATIVE_NET_ISTIO_RELEASE})"

  if [[ -z "${ISTIO_VERSION:-}" ]]; then
    readonly ISTIO_VERSION="stable"
  fi

#  if [[ -z "${NET_ISTIO_COMMIT:-}" ]]; then
#    NET_ISTIO_COMMIT=$(head -n 1 ${1} | grep "# Generated when HEAD was" | sed 's/^.* //')
#    echo "Got NET_ISTIO_COMMIT from ${1}: ${NET_ISTIO_COMMIT}"
#  fi
#
#  # TODO: remove this when all the net-istio.yaml in use contain a commit ID
#  if [[ -z "${NET_ISTIO_COMMIT:-}" ]]; then
#    NET_ISTIO_COMMIT="8102cd3d32f05be1c58260a9717d532a4a6d2f60"
#    echo "Hard coded NET_ISTIO_COMMIT: ${NET_ISTIO_COMMIT}"
#  fi
  LATEST_NET_ISTIO_RELEASE_VERSION=$(
  curl -L --silent "https://api.github.com/repos/knative/net-istio/releases" | grep '"tag_name"' \
    | cut -f2 -d: | sed "s/[^v0-9.]//g" | sort | tail -n1)

  # And checkout the setup script based on that commit.
  local NET_ISTIO_DIR=$(mktemp -d)
  (
    cd $NET_ISTIO_DIR \
      && git init \
      && git remote add origin https://github.com/knative-sandbox/net-istio.git \
      && git fetch --depth 1 origin $LATEST_NET_ISTIO_RELEASE_VERSION \
      && git checkout FETCH_HEAD
  )

  ISTIO_PROFILE="istio"
  if [[ -n "${KIND:-}" ]]; then
    ISTIO_PROFILE+="-kind"
  else
    ISTIO_PROFILE+="-ci"
  fi
#  if [[ $MESH -eq 0 ]]; then
#    ISTIO_PROFILE+="-no"
#  fi
  ISTIO_PROFILE+="-no-mesh"
  ISTIO_PROFILE+=".yaml"

  if [[ -n "${CLUSTER_DOMAIN:-}" ]]; then
    sed -ie "s/cluster\.local/${CLUSTER_DOMAIN}/g" ${NET_ISTIO_DIR}/third_party/istio-${ISTIO_VERSION}/${ISTIO_PROFILE}
  fi

  echo ">> Installing Istio"
  echo "Istio version: ${ISTIO_VERSION}"
  echo "Istio profile: ${ISTIO_PROFILE}"
  ${NET_ISTIO_DIR}/third_party/istio-${ISTIO_VERSION}/install-istio.sh ${ISTIO_PROFILE}

#  if [[ -n "${1:-}" ]]; then
#    echo ">> Installing net-istio"
#    echo "net-istio original YAML: ${1}"
#    # Create temp copy in which we replace knative-serving by the test's system namespace.
#    local YAML_NAME=$(mktemp -p $TMP_DIR --suffix=.$(basename "$1"))
#    sed "s/namespace: \"*${KNATIVE_DEFAULT_NAMESPACE}\"*/namespace: ${SYSTEM_NAMESPACE}/g" ${1} > ${YAML_NAME}
#    echo "net-istio patched YAML: $YAML_NAME"
#    ko apply -f "${YAML_NAME}" --selector=networking.knative.dev/ingress-provider=istio || return 1
#
#    CONFIGURE_ISTIO=${NET_ISTIO_DIR}/third_party/istio-${ISTIO_VERSION}/extras/configure-istio.sh
#    if [[ -f "$CONFIGURE_ISTIO" ]]; then
#      $CONFIGURE_ISTIO
#    else
#      echo "configure-istio.sh not found; skipping."
#    fi
#
#    UNINSTALL_LIST+=( "${YAML_NAME}" )
#  fi
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

    subheader "Installing eventing extension: sugar-controller (${eventing_version})"
    # install the sugar controller
    kubectl apply --filename https://storage.googleapis.com/knative-nightly/eventing/latest/eventing-sugar-controller.yaml
    wait_until_pods_running knative-eventing || return 1

  else
    start_release_knative_eventing "${eventing_version}"

    subheader "Installing eventing extension: sugar-controller (${eventing_version})"
    # install the sugar controller
    kubectl apply --filename https://storage.googleapis.com/knative-releases/eventing/previous/v${eventing_version}/eventing-sugar-controller.yaml
    wait_until_pods_running knative-eventing || return 1
  fi
}

# Create test resources and images
function test_setup() {
  echo ">> Uploading test images..."
  ${REPO_ROOT_DIR}/test/upload-test-images.sh || return 1
}
