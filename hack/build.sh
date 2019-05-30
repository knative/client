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

# Store for later
if [ -z "$1" ]; then
    ARGS=("")
else
    ARGS=("$@")
fi

set -eu

# Run build
run() {
  export GO111MODULE=on

  # Jump in/out project directory
  pushd $(basedir) >/dev/null 2>&1
  trap "popd >/dev/null 2>&1" EXIT

  if $(has_flag --help -h); then
    display_help
    exit 0
  fi

  # Switch on modules unconditionally
  export GO111MODULE=on

  if $(has_flag -u --update); then
    # Update dependencies
    update_deps
  fi

  # Run build
  go_build

  # Run tests
  if  $(has_flag --test -t) || ! $(has_flag --fast -f); then
    go_test
  fi

  if ! $(has_flag --fast -f); then
    # Format source code
    go_fmt

    # Generate docs
    generate_docs
  fi

  echo "────────────────────────────────────────────"
  ./kn version
}

go_fmt() {
  echo "🧹  Format"
  go fmt ./cmd/... ./pkg/...
}

go_build() {
  echo "🚧 Compile"
  source "./hack/build-flags.sh"
  go build -mod=vendor -ldflags "$(build_flags .)" -o ./kn ./cmd/...
}

go_test() {
  local test_output=$(mktemp /tmp/kn-client-test-output.XXXXXX)
  local red="[31m"
  local reset="[39m"

  echo "🧪  Test"
  set +e
  go test -v ./pkg/... >$test_output 2>&1
  local err=$?
  if [ $err -ne 0 ]; then
    echo "🔥 ${red}Failure${reset}"
    cat $test_output | sed -e "s/^.*\(FAIL.*\)$/$red\1$reset/"
    rm $test_output
    exit $err
  fi
  rm $test_output
}

update_deps() {
  echo "🕸️  Update"
  go mod vendor
}
generate_docs() {
  echo "📖 Docs"
  rm -rf "./docs/cmd"
  mkdir -p "./docs/cmd"
  go run "./hack/generate-docs.go" "."
}

# Dir where this script is located
basedir() {
    # Default is current directory
    local script=${BASH_SOURCE[0]}

    # Resolve symbolic links
    if [ -L $script ]; then
        if readlink -f $script >/dev/null 2>&1; then
            script=$(readlink -f $script)
        elif readlink $script >/dev/null 2>&1; then
            script=$(readlink $script)
        elif realpath $script >/dev/null 2>&1; then
            script=$(realpath $script)
        else
            echo "ERROR: Cannot resolve symbolic link $script"
            exit 1
        fi
    fi

    local dir=$(dirname "$script")
    local full_dir=$(cd "${dir}/.." && pwd)
    echo ${full_dir}
}

# Checks if a flag is present in the arguments.
has_flag() {
    filters="$@"
    for var in "${ARGS[@]}"; do
        for filter in $filters; do
          if [ "$var" = "$filter" ]; then
              echo 'true'
              return
          fi
        done
    done
    echo 'false'
}

# Display a help message.
display_help() {
    local command="${1:-}"
    cat <<EOT
Knative Client Build Script

Usage: $(basename $BASH_SOURCE) [... options ...]

with the following options:

-f  --fast                    Only build (without formatting, testing, code generation)
-t  --test                    Run tests even when used with --fast
-u  --update                  Update dependencies
-h  --help                    Display this help message
    --verbose                 Verbose script output (set -x)

You can add a symbolic link to this build script into your PATH so that it can be
called from everywhere. E.g.:

ln -s $(basedir)/hack/build.sh /usr/local/bin/kn_build.sh

Examples:

* Compile, format, tests, docs:  build.sh
* Compile only:                  build.sh --fast
* Compile with tests:            build.sh -f -t
EOT
}

if $(has_flag --verbose); then
    export PS4='+($(basename ${BASH_SOURCE[0]}):${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
    set -x
fi

run $*
