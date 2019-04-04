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

# This script runs the presubmit tests; it is started by prow for each PR.
# For convenience, it can also be executed manually.
# Running the script without parameters, or with the --all-tests
# flag, causes all tests to be executed, in the right order.
# Use the flags --build-tests, --unit-tests and --integration-tests
# to run a specific set of tests.

# Markdown linting failures don't show up properly in Gubernator resulting
# in a net-negative contributor experience.
export DISABLE_MD_LINTING=1
export GO111MODULE=on
source $(dirname $0)/../vendor/github.com/knative/test-infra/scripts/presubmit-tests.sh

# Checking licenses doesn't work yet with go mods.
# This is mostly the default runner but doesn't check them.

function build_tests() {
  local failed=0
  # Perform markdown build checks first
  markdown_build_tests || failed=1
  # For documentation PRs, just check the md files
  (( IS_DOCUMENTATION_PR )) && return ${failed}
  # Skip build test if there is no go code
  local go_pkg_dirs="$(go list ./...)"
  [[ -z "${go_pkg_dirs}" ]] && return ${failed}
  # Ensure all the code builds
  subheader "Checking that go code builds"
  go build -v ./... || failed=1
  # Get all build tags in go code (ignore /vendor)
  local tags="$(grep -r '// +build' . \
      | grep -v '^./vendor/' | cut -f3 -d' ' | sort | uniq | tr '\n' ' ')"
  if [[ -n "${tags}" ]]; then
    go test -run=^$ -tags="${tags}" ./... || failed=1
  fi
  if [[ -f ./hack/verify-codegen.sh ]]; then
    subheader "Checking autogenerated code is up-to-date"
    ./hack/verify-codegen.sh || failed=1
  fi
  subheader "Skipping license test"
  return ${failed}
}

main $@
