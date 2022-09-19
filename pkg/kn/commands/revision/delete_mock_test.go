/*
Copyright 2020 The Knative Authors

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
package revision

import (
	"errors"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/mock"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestRevisionDeletePruneAllMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	// Wait for delete event
	r.DeleteRevision("foo", mock.Any(), nil)
	r.DeleteRevision("bar", mock.Any(), nil)
	r.DeleteRevision("baz", mock.Any(), nil)

	revision1 := createMockRevisionWithParams("foo", "svc1", "1", "50", "")
	revision1.Labels[serving.RoutingStateLabelKey] = string(servingv1.RoutingStateReserve)
	revision2 := createMockRevisionWithParams("bar", "svc2", "1", "50", "")
	revision2.Labels[serving.RoutingStateLabelKey] = string(servingv1.RoutingStateReserve)
	revision3 := createMockRevisionWithParams("baz", "svc3", "1", "50", "")
	revision3.Labels[serving.RoutingStateLabelKey] = string(servingv1.RoutingStateReserve)
	revisionList := &servingv1.RevisionList{Items: []servingv1.Revision{*revision1, *revision2, *revision3}}
	r.ListRevisions(mock.Any(), revisionList, nil)

	output, err := executeRevisionCommand(client, "delete", "--prune-all")
	fmt.Println(output)
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "Revision", "deleted", "foo", "bar", "baz", "default"))
	r.Validate()
}

func TestRevisionDeleteCheckErrorForNotFoundRevisionsMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	r.DeleteRevision("foo", mock.Any(), nil)
	r.DeleteRevision("bar", mock.Any(), errors.New("revisions.serving.knative.dev \"bar\" not found"))
	r.DeleteRevision("baz", mock.Any(), errors.New("revisions.serving.knative.dev \"baz\" not found"))

	output, err := executeRevisionCommand(client, "delete", "foo", "bar", "baz")
	if err == nil {
		t.Fatal("Expected revision not found error, returned nil")
	}
	assert.Assert(t, util.ContainsAll(output, "'foo' deleted", "\"bar\" not found", "\"baz\" not found"))

	r.Validate()
}

func TestRevisionDeletePruneWithArgMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)
	// Recording:
	r := client.Recorder()
	r.DeleteRevision("foo", mock.Any(), nil)
	r.DeleteRevision("baz", mock.Any(), nil)
	r.DeleteRevision("bar", mock.Any(), nil)

	revision1 := createMockRevisionWithParams("foo", "svc1", "1", "50", "")
	revision1.Labels[serving.RoutingStateLabelKey] = string(servingv1.RoutingStateReserve)
	revision2 := createMockRevisionWithParams("bar", "svc1", "1", "50", "")
	revision2.Labels[serving.RoutingStateLabelKey] = string(servingv1.RoutingStateActive)
	revision3 := createMockRevisionWithParams("baz", "svc1", "1", "50", "")
	revision3.Labels[serving.RoutingStateLabelKey] = string(servingv1.RoutingStateActive)
	revisionList := &servingv1.RevisionList{Items: []servingv1.Revision{*revision1, *revision2, *revision3}}
	r.ListRevisions(mock.Any(), revisionList, nil)

	output, err := executeRevisionCommand(client, "delete", "--prune", "svc1")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "Revision", "deleted", "foo", "default"))
	assert.Assert(t, util.ContainsNone(output, "bar", "baz"))

}

func TestRevisionDeletePruneErrorFromArgMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)
	// Recording:
	r := client.Recorder()
	revisionList := &servingv1.RevisionList{Items: []servingv1.Revision{}}
	r.ListRevisions(mock.Any(), revisionList, nil)

	_, err := executeRevisionCommand(client, "delete", "--prune")
	assert.Error(t, err, "flag needs an argument: --prune")
}

func TestRevisionDeletePruneNoRevisionsMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()
	revisionList := &servingv1.RevisionList{Items: []servingv1.Revision{}}
	r.ListRevisions(mock.Any(), revisionList, nil)

	output, err := executeRevisionCommand(client, "delete", "--prune", "mysvc")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "unreferenced", "revisions", "found"))

	r.Validate()
}

func TestRevisionDeleteNoNameMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	_, err := executeRevisionCommand(client, "delete")
	assert.ErrorContains(t, err, "'kn revision delete' requires one or more revision name")

	r.Validate()

}

func TestRevisionDeletePruneAllNoRevisionsMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()
	revisionList := &servingv1.RevisionList{Items: []servingv1.Revision{}}
	r.ListRevisions(mock.Any(), revisionList, nil)

	output, err := executeRevisionCommand(client, "delete", "--prune-all")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "unreferenced", "revisions", "found"))

	r.Validate()
}

func TestRevisionDeletePruneAllErrorFromArgMock(t *testing.T) {
	// New mock client
	client := clientservingv1.NewMockKnServiceClient(t)
	// Recording:
	r := client.Recorder()
	revisionList := &servingv1.RevisionList{Items: []servingv1.Revision{}}
	r.ListRevisions(mock.Any(), revisionList, nil)

	_, err := executeRevisionCommand(client, "delete", "--prune-all", "mysvc")
	assert.Error(t, err, "'kn revision delete' with --prune-all flag requires no arguments")
}
