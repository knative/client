/*
Copyright 2023 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"fmt"
	"io"
)

var outWriter io.Writer

type Config struct {
	Quiet bool
	Out   io.Writer
}

func InitLogger(config Config) io.Writer {
	if config.Quiet {
		outWriter = io.Discard
	} else {
		outWriter = config.Out
	}

	return outWriter
}

func Fprintf(format string, args ...interface{}) {
	fmt.Fprintf(outWriter, format, args...)
}

func Fprintln(args ...interface{}) {
	fmt.Fprintln(outWriter, args...)
}
