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
      && git remote add origin https://github.com/knative-extensions/net-istio.git \
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
  kubectl apply -f ${NET_ISTIO_DIR}/third_party/istio-${ISTIO_VERSION}/${ISTIO_PROFILE%%.*}/istio.yaml

}

function knative_setup() {
  if is_ingress_class istio; then
    install_istio
  fi

  # Serving and Eventing 'latest' is based on branch context
  # On `main` it means 'nightly' manifests
  # On `release-*` it means corresponding `release-*` manifests
  # The described behavior is achieved through hack/library.sh start_latest_knative_serving()
  local serving_version="latest"
  local eventing_version="latest"

  # PR check variant triggering presubmit-integration-tests-latest-release
  # should check current client main branch code on latest available release
  # e.g. PR to 'main' on Serving and Eventing 1.3.0
  if [ "${LATEST_RELEASE}" == "true" ]; then
    serving_version="$(get_latest_release_version "knative" "serving")"
    eventing_version="$(get_latest_release_version "knative" "eventing")"
  fi

  header "Installing Knative Serving (${serving_version})"
  if [ "${serving_version}" = "latest" ]; then
    start_latest_knative_serving
  else
    # Serving and Net-Istio versions may differ on patch lvl
    start_knative_serving "https://storage.googleapis.com/knative-releases/serving/previous/v${serving_version}/serving-crds.yaml" \
      "https://storage.googleapis.com/knative-releases/serving/previous/v${serving_version}/serving-core.yaml" \
      "https://storage.googleapis.com/knative-releases/net-istio/previous/v$(get_latest_release_version "knative-extensions" "net-istio")/net-istio.yaml"
  fi

  if ! is_ingress_class istio; then
    kubectl patch configmap/config-network -n knative-serving \
      --type merge -p '{"data": {"ingress.class":"'${INGRESS_CLASS}'"}}'
  fi

  header "Installing Knative Eventing (${eventing_version})"
  if [ "${eventing_version}" = "latest" ]; then
    start_latest_knative_eventing
  else
    start_release_knative_eventing "${eventing_version}"
  fi
}

# Create test resources and images
function test_setup() {
  echo ">> Uploading test images..."
  ${REPO_ROOT_DIR}/test/upload-test-images.sh || return 1
}

# Retrieve latest version from given Knative repository tags
# On 'main' branch the latest released version is returned
# On 'release-x.y' branch the latest patch version for 'x.y.*' is returned
# Similar to hack/library.sh get_latest_knative_yaml_source()
function get_latest_release_version() {
    local org_name="$1"
    local repo_name="$2"
    local major_minor=""
    if is_release_branch; then
      local branch_name
      branch_name="$(current_branch)"
      major_minor="${branch_name##release-}"
    fi
    local version
    version="$(git ls-remote --tags --ref https://github.com/"${org_name}"/"${repo_name}".git \
      | grep "${major_minor}" \
      | cut -d '-' -f2 \
      | cut -d 'v' -f2 \
      | sort -Vr \
      | head -n 1)"
    echo "${version}"
}
