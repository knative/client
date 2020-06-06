// Copyright © 2019 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !windows

package plugin

// This file doesn't compile for Windows platform, therefor a second stat_windows.go is
// added with a no-op

import (
	"fmt"
	"os"
	"syscall"
)

func statFileOwner(fileInfo os.FileInfo) (uint32, uint32, error) {
	var sys *syscall.Stat_t
	var ok bool
	if sys, ok = fileInfo.Sys().(*syscall.Stat_t); !ok {
		return 0, 0, fmt.Errorf("cannot check owner/group of file %s", fileInfo.Name())
	}
	return sys.Uid, sys.Gid, nil
}
