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

# Temporary fix for iTerm issue https://gitlab.com/gnachman/iterm2/issues/7901
S=""
if [ -n "$ITERM_PROFILE" ]; then
  S=" "
fi
# Run build
run() {
  # Switch on modules unconditionally
  export GO111MODULE=on

  # Jump into project directory
  pushd $(basedir) >/dev/null 2>&1

  # Print help if requested
  if $(has_flag --help -h); then
    display_help
    exit 0
  fi

  if $(has_flag --watch -w); then
    watch
    # No exit, needs to be stopped with CTRL-C anyways
  fi

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

    # Check for license headers
    check_license

    # Auto generate cli docs
    generate_docs
  fi

  echo "────────────────────────────────────────────"
  ./kn version
}

go_fmt() {
  echo "🧹 ${S}Format"
  go fmt ./cmd/... ./pkg/...
}

go_build() {
  echo "🚧 Compile"
  source "./hack/build-flags.sh"
  go build -mod=vendor -ldflags "$(build_flags .)" -o kn ./cmd/...
}

go_test() {
  local test_output=$(mktemp /tmp/kn-client-test-output.XXXXXX)
  local red="[31m"
  local reset="[39m"

  echo "🧪 ${S}Test"
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

check_license() {
  echo "⚖️ ${S}License"
  local required_keywords=("Authors" "Apache License" "LICENSE-2.0")
  local extensions_to_check=("sh" "go" "yaml" "yml" "json")

  local check_output=$(mktemp /tmp/kn-client-licence-check.XXXXXX)
  for ext in "${extensions_to_check[@]}"; do
    find . -name "*.$ext" -a \! -path "./vendor/*" -a \! -path "./.*" -print0 |
      while IFS= read -r -d '' path; do
        for rword in "${required_keywords[@]}"; do
          if ! grep -q "$rword" "$path"; then
            echo "   $path" >> $check_output
          fi
        done
      done
  done
  if [ -s $check_output ]; then
    echo "🔥 No license header found in:"
    cat $check_output | sort | uniq
    echo "🔥 Please fix and retry."
    rm $check_output
    exit 1
  fi
  rm $check_output
}


update_deps() {
  echo "🕸️ Update"
  go mod vendor
}

generate_docs() {
  echo "📖 Docs"
  rm -rf "./docs/cmd"
  mkdir -p "./docs/cmd"
  go run "./hack/generate-docs.go" "."
}

watch() {
    local command="./hack/build.sh --fast"
    local fswatch_opts='-e "^\..*$" -o pkg cmd'
    if $(has_flag --test -t); then
      command="$command --test"
    fi
    if $(has_flag --verbose); then
      fswatch_opts="$fswatch_opts -v"
    fi
    set +e
    which fswatch >/dev/null 2>&1
    if [ $? -ne 0 ]; then
      local green="[32m"
      local reset="[39m"

      echo "🤷 Watch: Cannot find ${green}fswatch${reset}"
      echo "🌏 Please see ${green}http://emcrisostomo.github.io/fswatch/${reset} for installation instructions"
      exit 1
    fi
    set -e

    echo "🔁 Watch"
    fswatch $fswatch_opts | xargs -n1 -I{} $command
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
Knative client build script

Usage: $(basename $BASH_SOURCE) [... options ...]

with the following options:

-f  --fast                    Only compile (without formatting, testing, doc generation)
-t  --test                    Run tests when used with --fast or --watch
-u  --update                  Update dependencies before compiling
-w  --watch                   Watch for source changes and recompile in fast mode
-h  --help                    Display this help message
    --verbose                 More output
    --debug                   Debug information for this script (set -x)

You can add a symbolic link to this build script into your PATH so that it can be
called from everywhere. E.g.:

ln -s $(basedir)/hack/build.sh /usr/local/bin/kn_build.sh

Examples:

* Compile, format, tests, docs:  build.sh
* Compile only:                  build.sh --fast
* Compile with tests:            build.sh -f -t
* Automatice recompilation:      build.sh --watch
EOT
}

if $(has_flag --debug); then
    export PS4='+($(basename ${BASH_SOURCE[0]}):${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
    set -x
fi

run $*
