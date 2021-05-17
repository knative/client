// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

// KnRunResult holds command and result artifacts of a "kn" call
type KnRunResult struct {
	// Command line called
	CmdLine string
	// Standard output of command
	Stdout string
	// Standard error of command
	Stderr string
	// And extra dump informations in case of an unexpected error
	DumpInfo string
	// Error occurred during execution
	Error error
	// Was an error expected ?
	ErrorExpected bool
}

// KnRunResultCollector collects Kn run's results
type KnRunResultCollector struct {
	results    []KnRunResult
	extraDumps []string

	t      *testing.T
	knTest *KnTest
}

// NewKnRunResultCollector returns a new KnRunResultCollector
func NewKnRunResultCollector(t *testing.T, knTest *KnTest) *KnRunResultCollector {
	return &KnRunResultCollector{
		results:    []KnRunResult{},
		extraDumps: []string{},

		t:      t,
		knTest: knTest,
	}
}

// T returns the *testing.T object
func (c *KnRunResultCollector) T() *testing.T {
	return c.t
}

// KnTest returns the KnTest object
func (c *KnRunResultCollector) KnTest() *KnTest {
	return c.knTest
}

// AssertNoError helper to assert no error on result
func (c *KnRunResultCollector) AssertNoError(result KnRunResult) {
	c.results = append(c.results, result)
	if result.Error != nil {
		c.t.Logf("ERROR: %v", result.Stderr)
		c.t.FailNow()
	}
}

// AssertError helper to assert error on result
func (c *KnRunResultCollector) AssertError(result KnRunResult) {
	c.results = append(c.results, result)
	if result.Error == nil {
		c.t.Log("ERROR: Error expected but no error happened")
		c.t.FailNow()
	}
}

// AddDump adds extra dump information to the collector which is printed
// out if an error occurs
func (c *KnRunResultCollector) AddDump(kind string, name string, namespace string) {
	dumpInfo := extractDumpInfoWithName(kind, name, namespace)
	if dumpInfo != "" {
		c.extraDumps = append(c.extraDumps, dumpInfo)
	}
}

// DumpIfFailed logs if collector failed
func (c *KnRunResultCollector) DumpIfFailed() {
	if c.t.Failed() {
		c.Dump()
	}
}

// Dump prints out the collected output and logs
func (c *KnRunResultCollector) Dump() {
	c.t.Log(c.errorDetails())
}

// Private

func (c *KnRunResultCollector) errorDetails() string {
	var out = bytes.Buffer{}
	fmt.Fprintln(&out, "=== FAIL: =======================[[ERROR]]========================")
	c.printCommands(&out)
	var dumpInfos []string
	if len(c.results) > 0 {
		dumpInfo := c.results[len(c.results)-1].DumpInfo
		if dumpInfo != "" {
			dumpInfos = append(dumpInfos, dumpInfo)
		}
	}
	dumpInfos = append(dumpInfos, c.extraDumps...)
	for _, d := range dumpInfos {
		fmt.Fprintln(&out, "--------------------------[[DUMP]]-------------------------------")
		fmt.Fprint(&out, d)
	}

	fmt.Fprintln(&out, "=================================================================")
	return out.String()
}

func (c *KnRunResultCollector) printCommands(out io.Writer) {
	for i, result := range c.results {
		c.printCommand(out, result)
		if i < len(c.results)-1 {
			fmt.Fprintf(out, "┣━%s\n", separatorHeavy)
		}
	}
}

func (c *KnRunResultCollector) printCommand(out io.Writer, result KnRunResult) {
	fmt.Fprintf(out, "🦆 %s\n", result.CmdLine)
	for _, l := range strings.Split(result.Stdout, "\n") {
		fmt.Fprintf(out, "┃ %s\n", l)
	}
	if result.Stderr != "" {
		errorPrefix := "🔥"
		if result.ErrorExpected {
			errorPrefix = "︙"
		}
		for _, l := range strings.Split(result.Stderr, "\n") {
			fmt.Fprintf(out, "%s %s\n", errorPrefix, l)
		}
	}
}
