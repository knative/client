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

// Logger is the interface for logging, similar to the Uber's zap.Logger.
type Logger interface {
	// WithName returns a new Logger with the given name.
	WithName(name string) Logger

	// WithFields returns a new Logger with the given fields.
	WithFields(fields Fields) Logger

	// Debug logs a message at the debug level.
	Debug(args ...any)
	// Info logs a message at the info level.
	Info(args ...any)
	// Warn logs a message at the warn level.
	Warn(args ...any)
	// Error logs a message at the error level.
	Error(args ...any)
	// Fatal logs a message at the fatal level and then exit the program.
	Fatal(args ...any)

	// Debugf logs a message at the debug level using given format.
	Debugf(format string, args ...any)
	// Infof logs a message at the info level using given format.
	Infof(format string, args ...any)
	// Warnf logs a message at the warn level using given format.
	Warnf(format string, args ...any)
	// Errorf logs a message at the error level using given format.
	Errorf(format string, args ...any)
	// Fatalf logs a message at the fatal level using given format and then
	// exit the program.
	Fatalf(format string, args ...any)
}

// Fields is a map. It is used to add structured context to the logging output.
type Fields map[string]any
