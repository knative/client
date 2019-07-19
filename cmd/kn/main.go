// Copyright © 2018 The Knative Authors
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

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/knative/client/pkg/kn/core"
	"github.com/spf13/viper"
)

func init() {
	core.InitializeConfig()
}

var err error

func main() {
	defer cleanup()
	rand.Seed(time.Now().UnixNano())
	err = core.NewDefaultKnCommand().Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cleanup() {
	if err == nil {
		viper.WriteConfig()
	}
}
