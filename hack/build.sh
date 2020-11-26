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

source_dirs="cmd pkg test lib"

# Store for later
if [ -z "$1" ]; then
    ARGS=("")
else
    ARGS=("$@")
fi

set -eu

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
    # Build and test first
    go_build

    if $(has_flag --test -t); then
       go_test
    fi

    # Go in endless loop, to be stopped with CTRL-C
    watch
  fi

  # Fast mode: Only compile and maybe run test
  if $(has_flag --fast -f); then
    go_build

    if $(has_flag --test -t); then
       go_test
    fi
    exit 0
  fi

  # Run only tests
  if $(has_flag --test -t); then
    go_test
    exit 0
  fi

  # Run only codegen
  if $(has_flag --codegen -c); then
    codegen
    exit 0
  fi

  # Cross compile only
  if $(has_flag --all -x); then
    cross_build || (echo "✋ Cross platform build failed" && exit 1)
    exit 0
  fi

  # Default flow
  codegen
  go_build
  go_test

  echo "────────────────────────────────────────────"
  ./kn version
}


codegen() {
  # Update dependencies
  update_deps

  # Format source code and cleanup imports
  source_format

  # Check for license headers
  check_license

  # Auto generate cli docs
  generate_docs
}

go_fmt() {
  echo "🧹 ${S}Format"
  find $(echo $source_dirs) -name "*.go" -print0 | xargs -0 gofmt -s -w
}

# Run a go tool, get it first if necessary.
run_go_tool() {
  local tool=$2
  local install_failed=0
  if [ -z "$(which ${tool})" ]; then
    local temp_dir="$(mktemp -d)"
    pushd "${temp_dir}" > /dev/null 2>&1
    GOFLAGS="" go get "$1" || install_failed=1
    popd > /dev/null 2>&1
    rm -rf "${temp_dir}"
  fi
  (( install_failed )) && return ${install_failed}
  shift 2
  ${tool} "$@"
}


source_format() {
  set +e
  run_go_tool golang.org/x/tools/cmd/goimports goimports -w $(echo $source_dirs)
  find $(echo $source_dirs) -name "*.go" -print0 | xargs -0 gofmt -s -w
  set -e
}

go_build() {
  echo "🚧 Compile"
  go build -mod=vendor -ldflags "$(build_flags $(basedir))" -o kn ./cmd/...
}

go_test() {
  local test_output=$(mktemp /tmp/kn-client-test-output.XXXXXX)

  local red=""
  local reset=""
  # Use color only when a terminal is set
  if [ -t 1 ]; then
    red="[31m"
    reset="[39m"
  fi

  echo "🧪 ${X}Test"
  set +e
  go test -v ./cmd/... ./pkg/... >$test_output 2>&1
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
    find . -name "*.$ext" -a \! -path "./vendor/*" -a \! -path "./.*" -a \! -path "./third_party/*" -print0 |
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
  echo "🚒 Update"
  $(basedir)/hack/update-deps.sh
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
    fswatch $fswatch_opts | xargs -n1 -I{} sh -c "$command && echo 👌 OK"
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

cross_build() {
  local basedir=$(basedir)
  local ld_flags="$(build_flags $basedir)"
  local failed=0

  echo "⚔️ ${S}Compile"

  export CGO_ENABLED=0
  echo "   🐧 kn-linux-amd64"
  GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-amd64 ./cmd/... || failed=1
  echo "   💪 kn-linux-arm64"
  GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-arm64 ./cmd/... || failed=1
  echo "   🍏 kn-darwin-amd64"
  GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-darwin-amd64 ./cmd/... || failed=1
  echo "   🎠 kn-windows-amd64.exe"
  GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-windows-amd64.exe ./cmd/... || failed=1
  echo "   Z  kn-linux-s390x"
  GOOS=linux GOARCH=s390x go build -mod=vendor -ldflags "${ld_flags}" -o ./kn-linux-s390x ./cmd/... || failed=1

  return ${failed}
}

# Spaced fillers needed for certain emojis in certain terminals
S=""
X=""

# Calculate space fixing variables S and X
apply_emoji_fixes() {
  # Temporary fix for iTerm issue https://gitlab.com/gnachman/iterm2/issues/7901
  if [ -n "${ITERM_PROFILE:-}" ]; then
    S=" "
    # This issue has been fixed with iTerm2 3.3.7, so let's check for this
    # We can remove this code altogether if iTerm2 3.3.7 is in common usage everywhere
    if [ -n "${TERM_PROGRAM_VERSION}" ]; then
      args=$(echo $TERM_PROGRAM_VERSION | sed -e 's#[^0-9]*\([0-9]*\)[.]\([0-9]*\)[.]\([0-9]*\)\([0-9A-Za-z-]*\)#\1 \2 \3#')
      expanded=$(printf '%03d%03d%03d' $args)
      if [ $expanded -lt "003003007" ]; then
        X=" "
      fi
    fi
  fi
}

# Display a help message.
display_help() {
    local command="${1:-}"
    cat <<EOT
Knative client build script

Usage: $(basename $BASH_SOURCE) [... options ...]

with the following options:

-f  --fast                    Only compile (without dep update, formatting, testing, doc gen)
-t  --test                    Run tests when used with --fast or --watch
-c  --codegen                 Runs formatting, doc gen and update without compiling/testing
-w  --watch                   Watch for source changes and recompile in fast mode
-x  --all                     Only build cross platform binaries without code-generation/testing
-h  --help                    Display this help message
    --verbose                 More output
    --debug                   Debug information for this script (set -x)

You can add a symbolic link to this build script into your PATH so that it can be
called from everywhere. E.g.:

ln -s $(basedir)/hack/build.sh /usr/local/bin/kn_build.sh

Examples:

* Update deps, format, license check,
  gen docs, compile, test: ........... build.sh
* Compile only: ...................... build.sh --fast
* Run only tests: .................... build.sh --test
* Compile with tests: ................ build.sh -f -t
* Automatic recompilation: ........... build.sh --watch
* Build cross platform binaries: ..... build.sh --all
EOT
}

if $(has_flag --debug); then
    export PS4='+($(basename ${BASH_SOURCE[0]}):${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
    set -x
fi

# Shared funcs with CI
source $(basedir)/hack/build-flags.sh

# Fixe emoji labels for certain terminals
apply_emoji_fixes

run $*
