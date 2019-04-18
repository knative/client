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


ld_flags() {
   local dir=${1:-}
   version=$(get_version ${dir})
   now=$(date -u "+%Y-%m-%d %H:%M:%S")
   rev=$(git rev-parse --short HEAD)

   pkg="github.com/knative/client/pkg/kn/commands"
   echo "-X '${pkg}.BuildTime=$now' -X ${pkg}.Version=$version -X ${pkg}.GitRevision=$rev"
}

# Get version from local file
get_version() {
   local dir=${1:-}

   version=$(cat "$dir/NEXT_VERSION")
   date=$(date -u +%Y%m%d)
   echo "${version}-${date}"
}


