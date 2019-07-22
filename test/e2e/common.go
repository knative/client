// Copyright 2019 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

type runOpts struct {
	NoNamespace  bool
	AllowError   bool
	StderrWriter io.Writer
	StdoutWriter io.Writer
	StdinReader  io.Reader
	CancelCh     chan struct{}
	Redact       bool
}

// CreateTestNamespace creates and tests a namesspace creation invoking kubectl
func CreateTestNamespace(t *testing.T, namespace string) {
	kubectl := kubectl{t, Logger{}}
	out, err := kubectl.RunWithOpts([]string{"create", "namespace", namespace}, runOpts{})
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error executing 'kubectl create namespace' command. Error: %s", err.Error()))
	}

	expectedOutputRegexp := fmt.Sprintf("namespace?.+%s.+created", namespace)
	if !matchRegexp(t, expectedOutputRegexp, out) {
		t.Fatalf("Expected output incorrect, expecting to include:\n%s\n Instead found:\n%s\n", expectedOutputRegexp, out)
	}
}

// CreateTestNamespace deletes and tests a namesspace deletion invoking kubectl
func DeleteTestNamespace(t *testing.T, namespace string) {
	kubectl := kubectl{t, Logger{}}
	out, err := kubectl.RunWithOpts([]string{"delete", "namespace", namespace}, runOpts{})
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error executing 'kubectl delete namespace' command. Error: %s", err.Error()))
	}

	expectedOutputRegexp := fmt.Sprintf("namespace?.+%s.+deleted", namespace)
	if !matchRegexp(t, expectedOutputRegexp, out) {
		t.Fatalf("Expected output incorrect, expecting to include:\n%s\n Instead found:\n%s\n", expectedOutputRegexp, out)
	}
}

// Private functions

func runCLIWithOpts(cli string, args []string, opts runOpts, logger Logger) (string, error) {
	logger.Debugf("Running '%s'...\n", cmdCLIDesc(cli, args))

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command(cli, args...)
	cmd.Stderr = &stderr

	if opts.CancelCh != nil {
		go func() {
			select {
			case <-opts.CancelCh:
				cmd.Process.Signal(os.Interrupt)
			}
		}()
	}

	if opts.StdoutWriter != nil {
		cmd.Stdout = opts.StdoutWriter
	} else {
		cmd.Stdout = &stdout
	}

	cmd.Stdin = opts.StdinReader

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("Execution error: stderr: '%s' error: '%s'", stderr.String(), err)

		if !opts.AllowError {
			logger.Fatalf("Failed to successfully execute '%s': %v", cmdCLIDesc(cli, args), err)
		}
	}

	return stdout.String(), err
}

func cmdCLIDesc(cli string, args []string) string {
	return fmt.Sprintf("%s %s", cli, strings.Join(args, " "))
}

func matchRegexp(t *testing.T, matchingRegexp, actual string) bool {
	matched, err := regexp.MatchString(matchingRegexp, actual)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Failed to match regexp '%s'. Error: '%s'", matchingRegexp, err.Error()))
	}
	return matched
}
