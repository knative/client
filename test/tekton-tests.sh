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

# This script runs the integration tests with Tekton

# In order to run the tests locally against a standalone Kubernetes cluster
# and container registry, you must set the $CONTAINER_REGISTRY and $DOCKER_CONFIG_JSON
# environment variables and login to your registry of choice.

source $(dirname $0)/common.sh

# Add local dir to have access to built kn
export PATH=$PATH:${REPO_ROOT_DIR}

# Script entry point.
initialize $@

export TEKTON_VERSION=${TEKTON_VERSION:-v0.9.2}
export KN_E2E_NAMESPACE=tkn-kn

header "Running integration tests for Tekton"

# Install Tekton if not already installed
if [[ $(kubectl api-resources | grep -c tekton.dev) -eq 0 ]]; then
  kubectl apply -f https://github.com/tektoncd/pipeline/releases/download/${TEKTON_VERSION}/release.yaml
fi

if (( IS_PROW )); then
  # Configure Docker so that we can create a secret for GCR
  gcloud auth configure-docker
  gcloud auth print-access-token | docker login -u oauth2accesstoken --password-stdin https://gcr.io
  export CONTAINER_REGISTRY=gcr.io/${E2E_PROJECT_ID}/${E2E_BASE_NAME}-e2e-img/${RANDOM}
  export DOCKER_CONFIG_JSON=/root/.docker/config.json
fi

# Feed $KN_E2E_NAMESPACE and $CONTAINER_REGISTRY into yaml files
resource_dir=$(dirname $0)/resources/tekton
for file in kn-deployer-rbac kn-pipeline-resource; do
  sed -e "s#\${KN_E2E_NAMESPACE}#${KN_E2E_NAMESPACE}#" \
      -e "s#\${CONTAINER_REGISTRY}#${CONTAINER_REGISTRY}#" ${resource_dir}/${file}-template.yaml > ${resource_dir}/${file}.yaml
done

go_test_e2e -timeout=30m -tags=tekton ./test/e2e || fail_test

success
