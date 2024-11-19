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

function build_flags() {
  local now rev
  now="$(date -u '+%Y-%m-%d %H:%M:%S')"
  rev="$(git rev-parse --short HEAD)"
  local pkg="knative.dev/client/pkg/commands/version"
  local version="${TAG:-}"
  # Use vYYYYMMDD-local-<hash> for the version string, if not passed.
  if [[ -z "${version}" ]]; then
    # Get the commit, excluding any tags but keeping the "dirty" flag
    local commit
    commit="$(git describe --always --dirty --match '^$')"
    [[ -n "${commit}" ]] || abort "error getting the current commit"
    version="v$(date +%Y%m%d)-local-${commit}"
  fi
  export KN_BUILD_VERSION="${version}"
  export KN_BUILD_DATE="${now}"
  export KN_BUILD_GITREV="${rev}"
  KN_BUILD_LD_FLAGS="-X '${pkg}.BuildDate=${now}' \
  -X ${pkg}.Version=${version} \
  -X ${pkg}.GitRevision=${rev} \
  ${EXTERNAL_LD_FLAGS:-}"
}

build_flags
