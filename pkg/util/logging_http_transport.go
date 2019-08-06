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
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

type LoggingHttpTransport struct {
	transport http.RoundTripper
	stream    io.Writer
}

func NewLoggingTransport(transport http.RoundTripper) http.RoundTripper {
	return &LoggingHttpTransport{transport, nil}
}

func NewLoggingTransportWithStream(transport http.RoundTripper, s io.Writer) http.RoundTripper {
	return &LoggingHttpTransport{transport, s}
}

var SENSITIVE_HEADERS = map[string]bool{
	"Authorization":       true,
	"WWW-Authenticate":    true,
	"Cookie":              true,
	"Proxy-Authorization": true,
}

func (t *LoggingHttpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	stream := t.stream
	if stream == nil {
		stream = os.Stderr
	}
	reqCopy := *r
	reqCopy.Header = make(http.Header, len(r.Header))
	for k, v := range r.Header {
		sensitive := SENSITIVE_HEADERS[k]
		if sensitive {
			l := 0
			for _, h := range v {
				l += len(h)
			}
			reqCopy.Header.Set(k, strings.Repeat("*", l))
		} else {
			reqCopy.Header[k] = v
		}
	}
	reqBytes, _ := httputil.DumpRequestOut(&reqCopy, true)
	fmt.Fprintln(stream, "===== REQUEST =====")
	fmt.Fprintln(stream, string(reqBytes))
	resp, err := t.transport.RoundTrip(r)
	if err != nil {
		fmt.Fprintln(stream, "===== ERROR =====")
		fmt.Fprintln(stream, err)
	} else {
		respBytes, _ := httputil.DumpResponse(resp, true)
		fmt.Fprintln(stream, "===== RESPONSE =====")
		fmt.Fprintln(stream, string(respBytes))
		fmt.Fprintln(stream, " * * * * * *")
	}
	return resp, err
}
