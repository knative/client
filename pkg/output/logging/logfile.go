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

package logging

import (
	"context"
	"log"
	"os"
	"path"

	configdir "knative.dev/client/pkg/config/dir"
)

type logFileKey struct{}
type logFileCloserKey struct{}

type Closer func() error

// EnsureLogFile ensures that a log file is present in the context. If not, it
// creates one and adds it to the context. The file's Closer is also added to
// the context if not already present.
func EnsureLogFile(ctx context.Context) context.Context {
	var file *os.File
	if f := LogFileFrom(ctx); f == nil {
		file = createLogFile(ctx)
		ctx = WithLogFile(ctx, file)
	}
	var closer Closer
	if c := LogFileCloserFrom(ctx); c == nil {
		closer = file.Close
		ctx = WithLogFileCloser(ctx, closer)
	}
	return ctx
}

func LogFileFrom(ctx context.Context) *os.File {
	if f, ok := ctx.Value(logFileKey{}).(*os.File); ok {
		return f
	}
	return nil
}

func WithLogFile(ctx context.Context, f *os.File) context.Context {
	return context.WithValue(ctx, logFileKey{}, f)
}

func LogFileCloserFrom(ctx context.Context) Closer {
	if closer, ok := ctx.Value(logFileCloserKey{}).(Closer); ok {
		return closer
	}
	return nil
}

func WithLogFileCloser(ctx context.Context, closer Closer) context.Context {
	return context.WithValue(ctx, logFileCloserKey{}, closer)
}

func createLogFile(ctx context.Context) *os.File {
	cachePath := configdir.Cache(ctx)
	logPath := path.Join(cachePath, "last-exec.log.jsonl")
	if logFile, err := os.Create(logPath); err != nil {
		log.Fatal(err)
		return nil
	} else {
		return logFile
	}
}
