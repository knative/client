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

package logging_test

import (
	"context"
	"testing"

	"go.uber.org/zap/zapcore"
	"gotest.tools/v3/assert"
	pkgcontext "knative.dev/client/pkg/context"
	"knative.dev/client/pkg/output/logging"
)

func TestLogLevel(t *testing.T) {
	ctx := context.TODO()
	assert.Equal(t, zapcore.WarnLevel, logging.LogLevelFromContext(ctx))
	ctx = pkgcontext.WithTestingT(ctx, t)
	assert.Equal(t, zapcore.DebugLevel, logging.LogLevelFromContext(ctx))
	ctx = logging.WithLogLevel(ctx, zapcore.InfoLevel)
	assert.Equal(t, zapcore.InfoLevel, logging.LogLevelFromContext(ctx))
}
