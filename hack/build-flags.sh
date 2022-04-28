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
  local base="${1}"
  local now rev
  now="$(date -u '+%Y-%m-%d %H:%M:%S')"
  rev="$(git rev-parse --short HEAD)"
  local pkg="knative.dev/client/pkg/kn/commands/version"
  local version="${TAG:-}"
  local major_minor="$(echo "${version}" | cut -f1-2 -d. -n)"
  # Use vYYYYMMDD-local-<hash> for the version string, if not passed.
  if [[ -z "${version}" ]]; then
    # Get the commit, excluding any tags but keeping the "dirty" flag
    local commit
    commit="$(git describe --always --dirty --match '^$')"
    [[ -n "${commit}" ]] || abort "error getting the current commit"
    version="v$(date +%Y%m%d)-local-${commit}"
  fi
  # For Eventing and Serving versionings,
  # major and minor versions are the same as client version
  # patch version is from each technical version
  technical_version_serving=$(grep "knative.dev/serving " "${base}/go.mod" \
    | sed -s 's/.* \(v.[\.0-9]*\).*/\1/')
  technical_version_eventing=$(grep "knative.dev/eventing " "${base}/go.mod" \
    | sed -s 's/.* \(v.[\.0-9]*\).*/\1/')
  local version_serving version_eventing
  if [[ -n "${major_minor}" ]]; then
    version_serving=${major_minor}.$(echo ${technical_version_serving} |cut -f3 -d.)
    version_eventing=${major_minor}.$(echo ${technical_version_eventing} |cut -f3 -d.)
  else
    version_serving="${technical_version_serving}"
    version_eventing="${technical_version_eventing}"
  fi
  # Export as env variables to be used in `ko` OCI image build
  export KN_BUILD_VERSION="${version}"
  export KN_BUILD_DATE="${now}"
  export KN_BUILD_GITREV="${rev}"
  echo "-X '${pkg}.BuildDate=${now}' \
  -X ${pkg}.Version=${version} \
  -X ${pkg}.GitRevision=${rev} \
  -X ${pkg}.VersionServing=${version_serving} \
  -X ${pkg}.VersionEventing=${version_eventing}\
  ${EXTERNAL_LD_FLAGS:-}"
}
