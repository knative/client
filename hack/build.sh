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

dir=$(dirname "${BASH_SOURCE[0]}")
base=$(cd "$dir/.." && pwd)
source ${base}/hack/util/flags.sh

export GO111MODULE=on

echo "📋 Formatting"
go fmt ${base}/cmd/... ${base}/pkg/...
echo "🚧 Building"
go build -mod=vendor -ldflags "$(ld_flags ${base}/hack)" -o ${base}/kn ${base}/cmd/...
echo "🌞 Success"
${base}/kn version
