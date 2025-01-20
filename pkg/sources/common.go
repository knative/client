// Copyright © 2021 The Knative Authors
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

package sources

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"
)

// BuiltInSourcesGVKs returns the GVKs for built in sources
func BuiltInSourcesGVKs() []schema.GroupVersionKind {
	return []schema.GroupVersionKind{
		v1.SchemeGroupVersion.WithKind("ApiServerSource"),
		v1.SchemeGroupVersion.WithKind("ContainerSource"),
		v1.SchemeGroupVersion.WithKind("SinkBinding"),
		v1.SchemeGroupVersion.WithKind("PingSource"),
	}
}
