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

# Documentation about this script and how to use it can be found
# at https://github.com/knative/test-infra/tree/master/ci

source $(dirname $0)/../vendor/knative.dev/test-infra/scripts/release.sh
source $(dirname $0)/build-flags.sh

function build_release() {
  local ld_flags="$(build_flags $(dirname $0)/..)"
  local pkg="knative.dev/client/pkg/kn/commands"
  local version="${TAG}"
  # Use vYYYYMMDD-<hash>-local for the version string, if not passed.
  [[ -z "${version}" ]] && version="v${BUILD_TAG}-local"

  export GO111MODULE=on
  export CGO_ENABLED=0
  echo "🚧 🐧 Building for Linux"
  GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-amd64 ./cmd/...
  echo "🚧 🍏 Building for macOS"
  GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-darwin-amd64 ./cmd/...
  echo "🚧 🎠 Building for Windows"
  GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-windows-amd64.exe ./cmd/...
  echo "🚧 🐳 Building the container image"
  ko resolve --strict ${KO_FLAGS} -f config/ > kn-image-location.yaml
  ARTIFACTS_TO_PUBLISH="kn-darwin-amd64 kn-linux-amd64 kn-windows-amd64.exe kn-image-location.yaml"
  if type sha256sum >/dev/null 2>&1; then
    echo "🧮     Checksum:"
    sha256sum ${ARTIFACTS_TO_PUBLISH}
  fi
}

main $@
