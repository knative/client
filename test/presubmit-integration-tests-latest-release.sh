#!/usr/bin/env bash

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

# This script is used in Knative/test-infra as a custom prow job to run the
# integration tests against Knative Serving / Eventing of a specific version.

export KNATIVE_SERVING_VERSION="0.13.0"
export KNATIVE_EVENTING_VERSION="0.13.1"

$(dirname $0)/presubmit-tests.sh --integration-tests
