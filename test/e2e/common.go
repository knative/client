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
	"time"
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

var (
	e env
	k kn
)

const (
	KnDefaultTestImage string = "gcr.io/knative-samples/helloworld-go"
	MaxRetries         int    = 10
)

// Setup set up an enviroment for kn integration test returns the Teardown cleanup function
func Setup(t *testing.T) func(t *testing.T) {
	e = buildEnv(t)
	e.Namespace = fmt.Sprintf("%s%d", e.Namespace, namespaceCount)
	namespaceCount++
	k = kn{t, e.Namespace, Logger{}}
	CreateTestNamespace(t, e.Namespace)
	return Teardown
}

// Teaddown clean up
func Teardown(t *testing.T) {
	DeleteTestNamespace(t, e.Namespace)
}

// CreateTestNamespace creates and tests a namesspace creation invoking kubectl
func CreateTestNamespace(t *testing.T, namespace string) {
	logger := Logger{}
	kubectlCreateNamespace := func() (string, error) {
		kubectl := kubectl{t, logger}
		return kubectl.RunWithOpts([]string{"create", "namespace", namespace}, runOpts{AllowError: true})
	}

	expectedOutputRegexp := fmt.Sprintf("namespace?.+%s.+created", namespace)
	var (
		created bool
		retries int
		out     string
		err     error
	)

	for !created && retries < MaxRetries {
		out, err = kubectlCreateNamespace()
		if err != nil && retries >= MaxRetries {
			t.Fatalf("Failed creating test namespace:%s\n after %d retries", namespace, retries)
		} else {
			if err == nil {
				created = true
				break
			}

			retries++
			logger.Debugf("Could not create namespace, waiting 30s, and tring again: %d of %d\n", retries, MaxRetries)
			time.Sleep(30 * time.Second)
		}
	}

	// check that last output indeed show created namespace
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

// WaitForNamespaceDeleted wait until namespace is deleted
func WaitForNamespaceDeleted(t *testing.T, namespace string) {
	logger := Logger{}
	kubectlGetNamespace := func() string {
		kubectl := kubectl{t, logger}
		out, err := kubectl.RunWithOpts([]string{"get", "namespace"}, runOpts{})
		if err != nil {
			t.Fatalf(fmt.Sprintf("Error executing 'kubectl get namespace' command. Error: %s", err.Error()))
			return ""
		}

		return out
	}

	logger.Debugf("Waiting for namespace: %s to be deleted", namespace)
	deleted := false
	retries := 0
	for !deleted && retries < MaxRetries {
		output := kubectlGetNamespace()
		if !strings.Contains(output, namespace) {
			deleted = true
			break
		}

		time.Sleep(5000 * time.Millisecond)
		retries++
	}

	if !deleted && retries >= MaxRetries {
		t.Fatalf(fmt.Sprintf("Error deleting namespace %s, timed out", namespace))
	} else {
		logger.Debugf("Namespace: %s is deleted!\n", namespace)
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
