/*
 Copyright 2024 The Knative Authors

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

package sink

import (
	"strings"
)

var (
	// DefaultFlagName is a default command-line flag name.
	DefaultFlagName = "sink"
	// DefaultFlagShorthand is a default command-line flag shorthand.
	DefaultFlagShorthand = "s"
)

// Usage returns a usage text which can be used to define sink-like flag.
func Usage(fname string) string {
	flag := "--" + fname
	return "Addressable sink for events. " +
		"You can specify a broker, channel, Knative service, Kubernetes service or URI. " +
		"Examples: '" + flag + " broker:nest' for a broker 'nest', " +
		"'" + flag + " channel:pipe' for a channel 'pipe', " +
		"'" + flag + " ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', " +
		"'" + flag + " https://event.receiver.uri' for an HTTP URI, " +
		"'" + flag + " ksvc:receiver' or simply '" + flag + " receiver' for a Knative service 'receiver' in the current namespace, " +
		"'" + flag + " svc:receiver:mynamespace' for a Kubernetes service 'receiver' in the 'mynamespace' namespace, " +
		"'" + flag + " special.eventing.dev/v1alpha1/channels:pipe' for GroupVersionResource of v1alpha1 'pipe'. " +
		"If a prefix is not provided, it is considered as a Knative service in the current namespace."
}

// parseSink takes the string given by the user into the prefix, name and namespace of
// the object. If the user put a URI instead, the prefix is empty and the name
// is the whole URI.
func parseSink(sink string) (string, string, string) {
	parts := strings.SplitN(sink, ":", 3)
	switch {
	case len(parts) == 1:
		return knativeServiceShorthand, parts[0], ""
	case parts[0] == "http" || parts[0] == "https":
		return "", sink, ""
	case len(parts) == 3:
		return parts[0], parts[1], parts[2]
	default:
		return parts[0], parts[1], ""
	}
}
