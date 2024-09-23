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

import "go.uber.org/zap"

// ZapLogger is a Google' zap logger based logger.
type ZapLogger struct {
	*zap.SugaredLogger
}

func (z ZapLogger) WithName(name string) Logger {
	return &ZapLogger{
		SugaredLogger: z.SugaredLogger.Named(name),
	}
}

func (z ZapLogger) WithFields(fields Fields) Logger {
	a := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		a = append(a, k, v)
	}

	return &ZapLogger{
		SugaredLogger: z.SugaredLogger.With(a...),
	}
}
