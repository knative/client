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

package flags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

var negPrefix = "no-"

// AddBothBoolFlags adds the given flag in both `--foo` and `--no-foo` variants.
// If you do this, make sure you call ReconcileBoolFlags later to catch errors and
// set the relationship between the flag values.
func AddBothBoolFlags(f *pflag.FlagSet, p *bool, name, short string, value bool, usage string) {

	negativeName := negPrefix + name

	f.BoolVarP(p, name, short, value, usage)
	f.Bool(negativeName, !value, "do not "+usage)

	if value {
		err := f.MarkHidden(name)
		if err != nil {
			panic(err)
		}
	} else {
		err := f.MarkHidden(negativeName)
		if err != nil {
			panic(err)
		}
	}
}

// ReconcileBoolFlags sets the value of the all the "--foo" flags based on
// "--no-foo" if provided, and returns an error if both were provided.
func ReconcileBoolFlags(f *pflag.FlagSet) error {
	var err error
	f.Visit(func(flag *pflag.Flag) {
		// Return early from our comprehension
		if err != nil {
			return
		}
		// Walk the "no-" versions of the flags. Make sure we didn't set
		// both, and set the positive value to the opposite of the "no-"
		// value if it exists.
		if strings.HasPrefix(flag.Name, "no-") {
			positiveName := flag.Name[len(negPrefix):len(flag.Name)]
			positive := f.Lookup(positiveName)
			if positive.Changed {
				err = fmt.Errorf("only one of --%s and --%s may be specified",
					flag.Name, positiveName)
				return
			}
			var noValue bool
			noValue, err = strconv.ParseBool(flag.Value.String())
			if err != nil {
				return
			}
			err = positive.Value.Set(strconv.FormatBool(!noValue))
		}
	})
	return err
}
