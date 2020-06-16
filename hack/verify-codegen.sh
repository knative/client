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

set -o errexit
set -o nounset
set -o pipefail

source $(dirname $0)/../scripts/test-infra/library.sh

"${REPO_ROOT_DIR}"/hack/build.sh --codegen
if output="$(git status --porcelain)" && [ -z "$output" ]; then
  echo "${REPO_ROOT_DIR} is up to date."
else
  echo "ERROR: Modified files found:"
  git status --porcelain
  echo "ERROR: Diff"
  git diff
  echo "ERROR: ${REPO_ROOT_DIR} is out of date. Please run ./hack/build.sh -c and commit."
  exit 1
fi
