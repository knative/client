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
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
)

const (
	KnDefaultTestImage string        = "gcr.io/knative-samples/helloworld-go"
	MaxRetries         int           = 10
	RetrySleepDuration time.Duration = 5 * time.Second
)

var nsMutex sync.Mutex
var serviceMutex sync.Mutex
var serviceCount int
var namespaceCount int

type e2eTest struct {
	namespace string
	kn        kn
}

func NewE2eTest() (*e2eTest, error) {
	ns := nextNamespace()

	err := createNamespace(ns)
	if err != nil {
		return nil, err
	}
	err = waitForNamespaceCreated(ns)
	if err != nil {
		return nil, err
	}

	return &e2eTest{
		namespace: ns,
		kn:        kn{ns},
	}, nil
}

func nextNamespace() string {
	ns := os.Getenv("KN_E2E_NAMESPACE")
	if ns == "" {
		ns = "kne2etests"
	}
	return fmt.Sprintf("%s%d", ns, getNextNamespaceId())
}

func getNextNamespaceId() int {
	nsMutex.Lock()
	defer nsMutex.Unlock()
	current := namespaceCount
	namespaceCount++
	return current
}

func getNextServiceName(base string) string {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()
	current := serviceCount
	serviceCount++
	return base + strconv.Itoa(current)
}

// Teardown clean up
func (test *e2eTest) Teardown() error {
	return deleteNamespace(test.namespace)
}

// createNamespace creates and tests a namesspace creation invoking kubectl
func createNamespace(namespace string) error {
	expectedOutputRegexp := fmt.Sprintf("namespace?.+%s.+created", namespace)
	out, err := createNamespaceWithRetry(namespace, MaxRetries)
	if err != nil {
		return errors.Wrap(err, "could not create namespace "+namespace)
	}

	// check that last output indeed show created namespace
	matched, err := matchRegexp(expectedOutputRegexp, out)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("Expected output incorrect, expecting to include:\n%s\n Instead found:\n%s\n", expectedOutputRegexp, out)
	}
	return nil
}

// createNamespace deletes and tests a namesspace deletion invoking kubectl
func deleteNamespace(namespace string) error {
	kubectl := kubectl{namespace}
	out, err := kubectl.Run("delete", "namespace", namespace)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Cannot delete namespace %s", namespace))
	}

	expectedOutputRegexp := fmt.Sprintf("namespace?.+%s.+deleted", namespace)
	matched, err := matchRegexp(expectedOutputRegexp, out)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("Expected output incorrect, expecting to include:\n%s\n Instead found:\n%s\n", expectedOutputRegexp, out)
	}
	return nil
}

// WaitForNamespaceDeleted wait until namespace is deleted
func waitForNamespaceDeleted(namespace string) error {
	deleted := checkNamespace(namespace, false, MaxRetries)
	if !deleted {
		return fmt.Errorf("error deleting namespace %s, timed out after %d retries", namespace, MaxRetries)
	}
	return nil
}

// WaitForNamespaceCreated wait until namespace is created
func waitForNamespaceCreated(namespace string) error {
	created := checkNamespace(namespace, true, MaxRetries)
	if !created {
		return fmt.Errorf("error creating namespace %s, timed out after %d retries", namespace, MaxRetries)
	}
	return nil
}

// Private functions
func checkNamespace(namespace string, created bool, maxRetries int) bool {
	retries := 0
	for retries < MaxRetries {
		output, _ := kubectl{}.Run("get", "namespace")

		// check for namespace deleted
		if !created && !strings.Contains(output, namespace) {
			return true
		}

		// check for namespace created
		if created && strings.Contains(output, namespace) {
			return true
		}

		retries++
		time.Sleep(RetrySleepDuration)
	}

	return true
}

func createNamespaceWithRetry(namespace string, maxRetries int) (string, error) {
	var (
		retries int
		err     error
		out     string
	)

	for retries < maxRetries {
		out, err = kubectl{}.Run("create", "namespace", namespace)
		if err == nil {
			return out, nil
		}
		retries++
		time.Sleep(RetrySleepDuration)
	}

	return out, err
}

func matchRegexp(matchingRegexp, actual string) (bool, error) {
	matched, err := regexp.MatchString(matchingRegexp, actual)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("failed to match regexp '%s'", matchingRegexp))
	}
	return matched, nil
}

func currentDir(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal("Unable to read current dir:", err)
	}
	return dir
}
