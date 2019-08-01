// Copyright © 2018 The Knative Authors
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
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

type LoggingHttpTransport struct {
	transport http.RoundTripper
}

func NewLoggingTransport(transport http.RoundTripper) http.RoundTripper {
	return &LoggingHttpTransport{transport}
}

func (t *LoggingHttpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	reqBytes, _ := httputil.DumpRequestOut(r, true)
	fmt.Fprintln(os.Stderr, "===== REQUEST =====")
	fmt.Fprintln(os.Stderr, string(reqBytes))
	resp, err := t.transport.RoundTrip(r)
	respBytes, _ := httputil.DumpResponse(resp, true)
	fmt.Fprintln(os.Stderr, "===== RESPONSE =====")
	fmt.Fprintln(os.Stderr, string(respBytes))
	fmt.Fprintln(os.Stderr, " * * * * * *")
	return resp, err
}
