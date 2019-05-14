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

set -e -u

dir=$(dirname "${BASH_SOURCE[0]}")
base=$(cd "$dir/.." && pwd)
source ${base}/hack/util/flags.sh

ld_flags="$(ld_flags ${base}/hack)"
export GO111MODULE=on
export CGO_ENABLED=0
echo "🚧 🐧 Building for Linux"
GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ${base}/kn-darwin-amd64 ${base}/cmd/...
echo "🚧 🍏 Building for macOS"
GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ${base}/kn-linux-amd64 ${base}/cmd/...
echo "🚧 🎠 Building for Windows"
GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ${base}/kn-windows-amd64.exe ${base}/cmd/...

if type sha256sum >/dev/null 2>&1; then
  echo "🧮     Checksum:"
  pushd ${base} >/dev/null
  sha256sum kn-*-amd64*
  popd >/dev/null
fi

echo "🌞    Success"

