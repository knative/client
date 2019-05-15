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

source $(dirname $0)/../vendor/github.com/knative/test-infra/scripts/release.sh

function build_release() {
  local now="$(date -u '+%Y-%m-%d %H:%M:%S')"
  local rev="$(git rev-parse --short HEAD)"
  local pkg="github.com/knative/client/pkg/kn/commands"
  local version="${TAG}"
  # Use vYYYYMMDD-local-<hash> for the version string, if not passed.
  if [[ -z "${version}" ]]; then
    # Get the commit, excluding any tags but keeping the "dirty" flag
    local commit="$(git describe --always --dirty --match '^$')"
    [[ -n "${commit}" ]] || abort "error getting the current commit"
    version="v$(date +%Y%m%d)-local-${commit}"
  fi
  local ld_flags="-X '${pkg}.BuildDate=${now}' -X ${pkg}.Version=${version} -X ${pkg}.GitRevision=${rev}"

  export GO111MODULE=on
  export CGO_ENABLED=0
  echo "🚧 🐧 Building for Linux"
  GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-amd64 ./cmd/...
  echo "🚧 🍏 Building for macOS"
  GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-darwin-amd64 ./cmd/...
  echo "🚧 🎠 Building for Windows"
  GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-windows-amd64.exe ./cmd/...
  ARTIFACTS_TO_PUBLISH="kn-darwin-amd64 kn-linux-amd64 kn-windows-amd64.exe"
  if type sha256sum >/dev/null 2>&1; then
    echo "🧮     Checksum:"
    sha256sum ${ARTIFACTS_TO_PUBLISH}
  fi
}

main $@
