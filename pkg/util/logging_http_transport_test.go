// Copyright © 2019 The Knative Authors
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

package util

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"

	"gotest.tools/assert"
)

type dummyTransport struct {
	requestDump string
}

func (d *dummyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return nil, fmt.Errorf("dumping request: %v", err)
	}
	d.requestDump = string(dump)
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.0",
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

type errorTransport struct{}

func (d *errorTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, errors.New("This is an error")
}

func TestWritesRequestResponse(t *testing.T) {
	out := &bytes.Buffer{}
	dt := &dummyTransport{}
	transport := NewLoggingTransportWithStream(dt, out)

	body := "{this: is, the: body, of: [the, request]}"
	req, _ := http.NewRequest("POST", "http://example.com", strings.NewReader(body))
	nonRedacted := "this string will be logged"
	redacted := "this string will be redacted"
	req.Header.Add("non-redacted", nonRedacted)
	req.Header.Add(sensitiveRequestHeaders.List()[0], redacted)

	_, e := transport.RoundTrip(req)
	assert.NilError(t, e)
	s := out.String()
	assert.Assert(t, strings.Contains(s, "REQUEST"))
	assert.Assert(t, strings.Contains(s, "RESPONSE"))
	assert.Assert(t, strings.Contains(s, body))
	assert.Assert(t, strings.Contains(s, nonRedacted))
	assert.Assert(t, !strings.Contains(s, redacted))

	assert.Assert(t, strings.Contains(dt.requestDump, body))
	assert.Assert(t, strings.Contains(dt.requestDump, nonRedacted))
	assert.Assert(t, strings.Contains(dt.requestDump, redacted))
}

func TestElideAuthorizationHeader(t *testing.T) {
	out := &bytes.Buffer{}
	transport := NewLoggingTransportWithStream(&dummyTransport{}, out)
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Normal-Header", "la la normal text")
	req.Header.Set("Authorization", "Bearer: SECRET")
	_, e := transport.RoundTrip(req)
	assert.NilError(t, e)
	s := out.String()
	assert.Assert(t, strings.Contains(s, "REQUEST"))
	assert.Assert(t, strings.Contains(s, "Authorization"))
	assert.Assert(t, strings.Contains(s, "********"))
	assert.Assert(t, strings.Contains(s, "la la normal text"))
	assert.Assert(t, !strings.Contains(s, "SECRET"))
	assert.Assert(t, strings.Contains(s, "RESPONSE"))
}

func TestWritesRequestError(t *testing.T) {
	out := &bytes.Buffer{}
	transport := NewLoggingTransportWithStream(&errorTransport{}, out)
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	transport.RoundTrip(req)
	s := out.String()
	assert.Assert(t, strings.Contains(s, "REQUEST"))
	assert.Assert(t, strings.Contains(s, "ERROR"))
}
