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
# at https://github.com/knative/hack

source $(dirname $0)/../vendor/knative.dev/hack/release.sh
source $(dirname $0)/build-flags.sh

function build_release() {
  # Env var exported by hack/build-flags.sh
  local ld_flags="${KN_BUILD_LD_FLAGS:-}"

  export CGO_ENABLED=0
  echo "🚧 🐧 Building for Linux (amd64)"
  GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-amd64 ./cmd/...
  echo "🚧 💪 Building for Linux (arm64)"
  GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-arm64 ./cmd/...
  echo "🚧 🍏 Building for macOS"
  GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-darwin-amd64 ./cmd/...
  echo "🚧 🍎 Building for macOS (arm64)"
  GOOS=darwin GOARCH=arm64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-darwin-arm64 ./cmd/...
  echo "🚧 🎠 Building for Windows"
  GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-windows-amd64.exe ./cmd/...
  echo "🚧 Z  Building for Linux(s390x)"
  GOOS=linux GOARCH=s390x go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-s390x ./cmd/...
  echo "🚧 P  Building for Linux (ppc64le)"
  GOOS=linux GOARCH=ppc64le go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-ppc64le ./cmd/...
  echo "🚧 🐳 Building the container image"
  ko resolve ${KO_FLAGS} -f config/ > kn-image-location.yaml
  ARTIFACTS_TO_PUBLISH="kn-darwin-amd64 kn-darwin-arm64 kn-linux-amd64 kn-linux-arm64 kn-windows-amd64.exe kn-linux-s390x kn-linux-ppc64le kn-image-location.yaml"
  sha256sum ${ARTIFACTS_TO_PUBLISH} > checksums.txt
  ARTIFACTS_TO_PUBLISH="${ARTIFACTS_TO_PUBLISH} checksums.txt"
  echo "🧮     Checksum:"
  cat checksums.txt
}

main "$@"
