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

package tui

import (
	"io"
	"log"
	"os"
)

// safeguardBubbletea964 will safeguard the io.Reader by returning a new
// io.Reader that will prevent the
// https://github.com/charmbracelet/bubbletea/issues/964 issue.
//
// TODO: Remove this function once the issue is resolved.
func safeguardBubbletea964(in io.Reader) io.Reader {
	if in == nil {
		return in
	}
	if in == os.Stdin {
		// this is not a *os.File, so it will not try to do the epoll stuff
		return bubbletea964Input{Reader: in}
	}
	if f, ok := in.(*os.File); ok {
		if st, err := f.Stat(); err != nil {
			log.Fatal("unexpected: ", err)
		} else {
			if !st.Mode().IsRegular() {
				if st.Name() != os.DevNull {
					log.Println("WARN: non-regular file given as input: ",
						st.Name(), " (mode: ", st.Mode(),
						"). Using `nil` as input.")
				}
				return nil
			}
		}
	}
	return in
}

type bubbletea964Input struct {
	io.Reader
}
