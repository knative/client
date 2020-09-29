#!/usr/bin/env bash

# Copyright 2020 The Knative Authors
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

readonly ROOT_DIR=$(dirname $0)/..
source ${ROOT_DIR}/scripts/test-infra/library.sh

set -o errexit
set -o nounset
set -o pipefail

cd ${ROOT_DIR}

VERSION="release-0.18"

# The list of dependencies that we track at HEAD and periodically
# float forward in this repository.
FLOATING_DEPS=(
  "knative.dev/pkg@${VERSION}"
  "knative.dev/networking@${VERSION}"
  "knative.dev/serving@${VERSION}"
  "knative.dev/eventing@${VERSION}"
)

# Parse flags to determine any we should pass to dep.
GO_GET=0
while [[ $# -ne 0 ]]; do
  parameter=$1
  case ${parameter} in
    --upgrade) GO_GET=1 ;;
    *) abort "unknown option ${parameter}" ;;
  esac
  shift
done
readonly GO_GET

if (( GO_GET )); then
  go get -d ${FLOATING_DEPS[@]}
  "${ROOT_DIR}/scripts/test-infra/update-test-infra.sh" --update --ref "release-0.18"
fi


# Prune modules.
go mod tidy
go mod vendor

# Cleanup
find "${ROOT_DIR}/vendor" \( -name "OWNERS" -o -name "*_test.go" \) -print0 | xargs -0 rm -f

export GOFLAGS=-mod=vendor
update_licenses "${ROOT_DIR}/third_party/VENDOR-LICENSE" "./..."

remove_broken_symlinks "${ROOT_DIR}/vendor"
