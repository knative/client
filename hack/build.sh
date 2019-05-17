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

source $(dirname $0)/../vendor/github.com/knative/test-infra/scripts/library.sh
cd ${REPO_ROOT_DIR}

set -o pipefail
set -eu

export GO111MODULE=on

echo "📋 Formatting"
go fmt ./cmd/... ./pkg/...
echo "🚧 Building"
./hack/release.sh --nopublish --skip-tests
# TODO(adrcunha): Use IS_* once knative/test-infra#793 is available here.
case "${OSTYPE}" in
  darwin*) ln -s kn-darwin-amd64 kn ;;
  linux*) ln -s kn-linux-amd64 kn ;;
  msys*) ln -s kn-windows-amd64.exe kn ;;
  *) abort "Unknown OS" ;;
esac
echo "🌞 Success"

./hack/generate-docs.sh
./kn version
