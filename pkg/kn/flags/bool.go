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
	"unicode"
	"unicode/utf8"

	"github.com/spf13/pflag"
)

var (
	negPrefix        = "no-"
	deprecatedPrefix = "DEPRECATED:"
)

// AddBothBoolFlagsUnhidden is just like AddBothBoolFlags but shows both flags.
func AddBothBoolFlagsUnhidden(f *pflag.FlagSet, p *bool, name, short string, value bool, usage string) {

	negativeName := negPrefix + name

	f.BoolVarP(p, name, short, value, usage)
	f.Bool(negativeName, !value, InvertUsage(usage))
}

// AddBothBoolFlags adds the given flag in both `--foo` and `--no-foo` variants.
// If you do this, make sure you call ReconcileBoolFlags later to catch errors
// and set the relationship between the flag values. Only the flag that does the
// non-default behavior is visible; the other is hidden.
func AddBothBoolFlags(f *pflag.FlagSet, p *bool, name, short string, value bool, usage string) {
	AddBothBoolFlagsUnhidden(f, p, name, short, value, usage)
	negativeName := negPrefix + name
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
// "--no-foo" if provided, and returns an error if both were provided or an
// explicit value of false was provided to either (as that's confusing).
func ReconcileBoolFlags(f *pflag.FlagSet) error {
	var err error
	f.VisitAll(func(flag *pflag.Flag) {
		// Return early from our comprehension
		if err != nil {
			return
		}

		// handle async flag
		if flag.Name == "async" && flag.Changed {
			if f.Lookup("wait").Changed || f.Lookup("no-wait").Changed {
				err = fmt.Errorf("only one of (DEPRECATED) --async, --wait and --no-wait may be specified")
				return
			}
			err = checkExplicitFalse(flag, "wait")
			if err != nil {
				return
			}
			f.Lookup("no-wait").Value.Set("true")
		}

		// Walk the "no-" versions of the flags. Make sure we didn't set
		// both, and set the positive value to the opposite of the "no-"
		// value if it exists.
		if strings.HasPrefix(flag.Name, negPrefix) {
			positiveName := flag.Name[len(negPrefix):]
			positive := f.Lookup(positiveName)
			// Non-paired flag, or wrong types
			if positive == nil || positive.Value.Type() != "bool" || flag.Value.Type() != "bool" {
				return
			}
			if flag.Changed {
				if positive.Changed {
					err = fmt.Errorf("only one of --%s and --%s may be specified",
						flag.Name, positiveName)
					return
				}
				err = checkExplicitFalse(flag, positiveName)
				if err != nil {
					return
				}
				err = positive.Value.Set("false")
			} else {
				err = checkExplicitFalse(positive, flag.Name)
			}
		}

	})
	return err
}

// checkExplicitFalse returns an error if the flag was explicitly set to false.
func checkExplicitFalse(f *pflag.Flag, betterFlag string) error {
	if !f.Changed {
		return nil
	}
	val, err := strconv.ParseBool(f.Value.String())
	if err != nil {
		return err
	}
	if !val {
		return fmt.Errorf("use --%s instead of providing \"%s\" to --%s",
			betterFlag, f.Value.String(), f.Name)
	}
	return nil
}

// FirstCharToLower converts first char in given string to lowercase
func FirstCharToLower(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

// InvertUsage inverts the usage string with prefix "Do not"
func InvertUsage(usage string) string {
	return "Do not " + FirstCharToLower(usage)
}
