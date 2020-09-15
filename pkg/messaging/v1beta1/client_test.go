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
package v1beta1

import (
	"testing"

	"gotest.tools/assert"

	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
)

func TestBuiltInChannelGVks(t *testing.T) {
	gvks := BuiltInChannelGVKs()
	for _, each := range gvks {
		assert.DeepEqual(t, each.GroupVersion(), messagingv1beta1.SchemeGroupVersion)
	}
	assert.Equal(t, len(gvks), 1)
}
