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
	"os/exec"
	"strings"
)

const (
	separatorHeavy = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	separatorLight = "╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍╍"
)

// Kn type
type Kn struct {
	namespace string
}

// NewKn object
func NewKn() Kn {
	return Kn{}
}

// RunNoNamespace the 'kn' CLI with args but no namespace
func (k Kn) RunNoNamespace(args ...string) KnRunResult {
	return RunKn("", args)
}

// Run the 'kn' CLI with args
func (k Kn) Run(args ...string) KnRunResult {
	return RunKn(k.namespace, args)
}

// Namespace that this Kn instance uses
func (k Kn) Namespace() string {
	return k.namespace
}

// Kubectl type
type Kubectl struct {
	namespace string
}

// New Kubectl object
func NewKubectl(namespace string) Kubectl {
	return Kubectl{
		namespace: namespace,
	}
}

// Run the 'kubectl' CLI with args
func (k Kubectl) Run(args ...string) (string, error) {
	return RunKubectl(k.namespace, args...)
}

// Namespace that this Kubectl instance uses
func (k Kubectl) Namespace() string {
	return k.namespace
}

// Public functions

// RunKn runs "kn" in a given namespace
func RunKn(namespace string, args []string) KnRunResult {
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	stdout, stderr, err := runCli("kn", args)
	result := KnRunResult{
		CmdLine: cmdCLIDesc("kn", args),
		Stdout:  stdout,
		Stderr:  stderr,
		Error:   err,
	}
	if err != nil {
		command := args[0]
		if command == "source" && len(args) > 1 {
			command = "source " + args[1]
			args = args[1:]
		}
		result.DumpInfo = extractDumpInfo(command, args, namespace)
	}
	return result
}

// RunKubectl runs "kubectl" in a given namespace
func RunKubectl(namespace string, args ...string) (string, error) {
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	stdout, stderr, err := runCli("kubectl", args)
	if err != nil {
		return stdout, fmt.Errorf("stderr: %s: %w", stderr, err)
	}
	return stdout, nil
}

// Private

func runCli(cli string, args []string) (string, string, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command(cli, args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	cmd.Stdin = nil

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

type dumpFunc func(namespace string, args []string) string

// Dump handler for specific commands ("service", "revision") which should add extra infos
// Relies on that argv[1] is the command and argv[3] is the name of the object
var dumpHandlers = map[string]dumpFunc{
	"service":          dumpService,
	"revision":         dumpRevision,
	"route":            dumpRoute,
	"trigger":          dumpTrigger,
	"source apiserver": dumpApiServerSource,
}

func extractDumpInfoWithName(command string, name string, namespace string) string {
	return extractDumpInfo(command, []string{command, "", name}, namespace)
}

func extractDumpInfo(command string, args []string, namespace string) string {
	dumpHandler := dumpHandlers[command]
	if dumpHandler != nil {
		return dumpHandler(namespace, args)
	}
	return ""
}

func dumpService(namespace string, args []string) string {
	// For list like operation we don't have a name
	if len(args) < 3 || args[2] == "" {
		return ""
	}
	name := args[2]
	var buffer bytes.Buffer

	// Service info
	appendResourceInfo(&buffer, "ksvc", name, namespace)
	fmt.Fprintf(&buffer, "%s\n", separatorHeavy)
	// Service's configuration
	appendResourceInfo(&buffer, "configuration", name, namespace)
	fmt.Fprintf(&buffer, "%s\n", separatorHeavy)
	// Get all revisions for this service
	appendResourceInfoWithNameSelector(&buffer, "revision", name, namespace, "serving.knative.dev/service")
	// Get all routes for this service
	appendResourceInfoWithNameSelector(&buffer, "route", name, namespace, "serving.knative.dev/service")
	return buffer.String()
}

func dumpRevision(namespace string, args []string) string {
	return simpleDump("revision", args, namespace)
}

func dumpRoute(namespace string, args []string) string {
	return simpleDump("route", args, namespace)
}

func dumpApiServerSource(namespace string, args []string) string {
	return simpleDump("apiserversource", args, namespace)
}

func dumpTrigger(namespace string, args []string) string {
	return simpleDump("trigger", args, namespace)
}

func simpleDump(kind string, args []string, namespace string) string {
	if len(args) < 3 || args[2] == "" {
		return ""
	}

	var buffer bytes.Buffer
	appendResourceInfo(&buffer, kind, args[2], namespace)
	return buffer.String()
}

func appendResourceInfo(buffer *bytes.Buffer, kind string, name string, namespace string) {
	appendResourceInfoWithNameSelector(buffer, kind, name, namespace, "")
}

func appendResourceInfoWithNameSelector(buffer *bytes.Buffer, kind string, name string, namespace string, selector string) {
	var extra string
	argsDescribe := []string{"describe", kind}
	argsGet := []string{"get", "-oyaml", kind}

	if selector != "" {
		labelArg := fmt.Sprintf("%s=%s", selector, name)
		argsDescribe = append(argsDescribe, "--selector", labelArg)
		argsGet = append(argsGet, "--selector", labelArg)
		extra = fmt.Sprintf(" --selector %s", labelArg)
	} else {
		argsDescribe = append(argsDescribe, name)
		argsGet = append(argsGet, name)
		extra = ""
	}

	out, err := RunKubectl(namespace, argsDescribe...)
	appendCLIOutput(buffer, fmt.Sprintf("kubectl describe %s %s --namespace %s%s", kind, name, namespace, extra), out, err)
	fmt.Fprintf(buffer, "%s\n", separatorLight)
	out, err = RunKubectl(namespace, argsGet...)
	appendCLIOutput(buffer, fmt.Sprintf("kubectl get %s %s --namespace %s -oyaml%s", kind, name, namespace, extra), out, err)
}

func appendCLIOutput(buffer *bytes.Buffer, desc string, out string, err error) {
	buffer.WriteString(fmt.Sprintf("==== %s\n", desc))
	if err != nil {
		buffer.WriteString(fmt.Sprintf("%s: %v\n", "!!!! ERROR", err))
	}
	buffer.WriteString(out)
}

func cmdCLIDesc(cli string, args []string) string {
	return fmt.Sprintf("%s %s", cli, strings.Join(args, " "))
}
