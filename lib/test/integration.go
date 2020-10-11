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

package test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	MaxRetries         int           = 10
	RetrySleepDuration time.Duration = 5 * time.Second
)

var nsMutex sync.Mutex
var serviceMutex sync.Mutex
var serviceCount int
var namespaceCount int

// KnTest type
type KnTest struct {
	namespace string
	kn        Kn
}

// NewKnTest creates a new KnTest object
func NewKnTest() (*KnTest, error) {
	ns := ""
	// try next 20 namespace before giving up creating a namespace if it already exists
	for i := 0; i < 20; i++ {
		ns = NextNamespace()
		err := CreateNamespace(ns)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "AlreadyExists") {
			continue
		} else {
			return nil, err
		}
	}

	err := WaitForNamespaceCreated(ns)
	if err != nil {
		return nil, err
	}

	return &KnTest{
		namespace: ns,
		kn:        Kn{ns},
	}, nil
}

// Teardown clean up
func (test *KnTest) Teardown() error {
	return DeleteNamespace(test.namespace)
}

// Kn object used by this KnTest
func (test *KnTest) Kn() Kn {
	return test.kn
}

// Namespace used by the test
func (test *KnTest) Namespace() string {
	return test.namespace
}

// Public functions

// NextNamespace return the next unique namespace
func NextNamespace() string {
	ns := os.Getenv("KN_E2E_NAMESPACE")
	if ns == "" {
		ns = "kne2etests"
	}
	return fmt.Sprintf("%s%d", ns, GetNextNamespaceId())
}

// GetNextNamespaceId return the next unique ID for the next namespace
func GetNextNamespaceId() int {
	nsMutex.Lock()
	defer nsMutex.Unlock()
	current := namespaceCount
	namespaceCount++
	return current
}

// GetNextServiceName return the name for the next namespace
func GetNextServiceName(base string) string {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()
	current := serviceCount
	serviceCount++
	return base + strconv.Itoa(current)
}

// CreateNamespace creates and tests a namespace creation invoking kubectl
func CreateNamespace(namespace string) error {
	expectedOutputRegexp := fmt.Sprintf("namespace?.+%s.+created", namespace)
	out, err := createNamespaceWithRetry(namespace, MaxRetries)
	if err != nil {
		return fmt.Errorf("could not create namespace %s: %w", namespace, err)
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

// DeleteNamespace deletes and tests a namespace deletion invoking kubectl
func DeleteNamespace(namespace string) error {
	kubectl := Kubectl{namespace}
	out, err := kubectl.Run("delete", "namespace", namespace)
	if err != nil {
		return fmt.Errorf("Cannot delete namespace %s: %w", namespace, err)
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
func WaitForNamespaceDeleted(namespace string) error {
	deleted := checkNamespace(namespace, false, MaxRetries)
	if !deleted {
		return fmt.Errorf("error deleting namespace %s, timed out after %d retries", namespace, MaxRetries)
	}
	return nil
}

// WaitForNamespaceCreated wait until namespace is created
func WaitForNamespaceCreated(namespace string) error {
	created := checkNamespace(namespace, true, MaxRetries)
	if !created {
		return fmt.Errorf("error creating namespace %s, timed out after %d retries", namespace, MaxRetries)
	}
	return nil
}

func CurrentDir(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal("Unable to read current dir:", err)
	}
	return dir
}

// Private functions

func checkNamespace(namespace string, created bool, maxRetries int) bool {
	retries := 0
	for retries < MaxRetries {
		output, _ := Kubectl{}.Run("get", "namespace")

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
		out, err = Kubectl{}.Run("create", "namespace", namespace)
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
		return false, fmt.Errorf("failed to match regexp %q: %w", matchingRegexp, err)
	}
	return matched, nil
}
