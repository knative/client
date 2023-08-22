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

package v1beta1

import (
	"context"
	"testing"

	"knative.dev/serving/pkg/apis/serving/v1beta1"
)

func TestMockKnClient(t *testing.T) {

	client := NewMockKnServiceClient(t)

	recorder := client.Recorder()

	// Record all services
	recorder.GetDomainMapping("hello.foo.bar", &v1beta1.DomainMapping{}, nil)
	recorder.CreateDomainMapping(&v1beta1.DomainMapping{}, nil)
	recorder.DeleteDomainMapping("hello.foo.bar", nil)
	recorder.UpdateDomainMapping(&v1beta1.DomainMapping{}, nil)

	recorder.GetDomainMapping("hello.foo.bar", &v1beta1.DomainMapping{}, nil)
	recorder.UpdateDomainMapping(&v1beta1.DomainMapping{}, nil)
	recorder.ListDomainMappings(&v1beta1.DomainMappingList{}, nil)

	// Call all services
	ctx := context.Background()
	client.GetDomainMapping(ctx, "hello.foo.bar")
	client.CreateDomainMapping(ctx, &v1beta1.DomainMapping{})
	client.DeleteDomainMapping(ctx, "hello.foo.bar")
	client.UpdateDomainMapping(ctx, &v1beta1.DomainMapping{})
	client.UpdateDomainMappingWithRetry(ctx, "hello.foo.bar", func(origDomain *v1beta1.DomainMapping) (*v1beta1.DomainMapping, error) {
		return origDomain, nil
	}, 10)
	client.ListDomainMappings(ctx)

	// Validate
	recorder.Validate()
}
