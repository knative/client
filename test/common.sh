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

source $(dirname $0)/../scripts/test-infra/e2e-tests.sh

function cluster_setup() {
  header "Building client"
  ${REPO_ROOT_DIR}/hack/build.sh -f || return 1
}

function knative_setup() {
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

    # install the sugar controller
    kubectl apply --filename https://storage.googleapis.com/knative-nightly/eventing/latest/eventing-sugar-controller.yaml
    wait_until_pods_running knative-eventing || return 1

  else
    start_release_knative_eventing "${eventing_version}"

    # install the sugar controller
    kubectl apply --filename https://storage.googleapis.com/knative-releases/eventing/previous/${eventing_version}/eventing-sugar-controller.yaml
    wait_until_pods_running knative-eventing || return 1
  fi
}
