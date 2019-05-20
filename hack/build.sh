#!/bin/bash

# Copyright 2018 The Knative Authors
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

set -o pipefail
set -eu

# Run build
run() {
  export GO111MODULE=on

  go_fmt
  go_build
  generate_docs

  echo "🌞 Success"

  $(basedir)/kn version
}

go_fmt() {
  local base=$(basedir)
  echo "📋 Formatting"
  go fmt "${base}/cmd/..." "${base}/pkg/..."
}

go_build() {
  local base=$(basedir)
  echo "🚧 Building"
  source "${base}/hack/build-flags.sh"
  go build -mod=vendor -ldflags "$(build_flags ${base})" -o ${base}/kn ${base}/cmd/...
}

generate_docs() {
  local base=$(basedir)
  echo "📑 Generating docs"
  rm -rf "${base}/docs/cmd"
  mkdir -p "${base}/docs/cmd"

  go run "${base}/hack/generate-docs.go" "${base}"
}

basedir() {
  dir=$(dirname "${BASH_SOURCE[0]}")
  base=$(cd "$dir/.." && pwd)
  echo ${base}
}

run $*