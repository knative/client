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

set -e -x -u

VERSION=0.0.0
BUILDTIME=`date -u +%Y%m%d.%H%M%S`

go fmt ./cmd/... ./pkg/...

go build -ldflags "-X github.com/knative/client/pkg/kn/commands.Version=$VERSION.$BUILDTIME" ./cmd/...

./kn version

echo "Success"
