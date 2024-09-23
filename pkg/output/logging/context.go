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
	"os"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	pkgcontext "knative.dev/client/pkg/context"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/environment"
	"knative.dev/client/pkg/output/term"
	"knative.dev/pkg/logging"
)

// ErrCallEnsureLoggerFirst is returned when LoggerFrom() is called before EnsureLogger().
var ErrCallEnsureLoggerFirst = errors.New("call EnsureLogger() before LoggerFrom() method")

// EnsureLogger ensures that a logger is attached to the context. The returned
// context will have a logger attached to it. Given fields will be added to the
// logger, either new or existing.
func EnsureLogger(ctx context.Context, fields ...Fields) context.Context {
	z, err := zapLoggerFrom(ctx)
	if errors.Is(err, ErrCallEnsureLoggerFirst) {
		ctx = EnsureLogFile(ctx)
		z = setupLogging(ctx)
	}
	l := &ZapLogger{SugaredLogger: z}
	for _, f := range fields {
		l = l.WithFields(f).(*ZapLogger)
	}
	return WithLogger(ctx, l)
}

// LoggerFrom returns the logger from the context. If EnsureLogger() was not
// called before, it will panic.
func LoggerFrom(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerKey).(Logger); ok {
		return l
	}
	z, err := zapLoggerFrom(ctx)
	if err != nil {
		fatal(err)
	}

	return &ZapLogger{z}
}

// WithLogger attaches the given logger to the context.
func WithLogger(ctx context.Context, l Logger) context.Context {
	if z, ok := l.(*ZapLogger); ok {
		return logging.WithLogger(ctx, z.SugaredLogger)
	}
	return context.WithValue(ctx, loggerKey, l)
}

var loggerKey = struct{}{}

func zapLoggerFrom(ctx context.Context) (*zap.SugaredLogger, error) {
	l := logging.FromContext(ctx)
	if l.Desugar().Name() == "fallback" {
		return nil, ErrCallEnsureLoggerFirst
	}
	return l, nil
}

func setupLogging(ctx context.Context) *zap.SugaredLogger {
	var logger *zap.Logger
	if t := pkgcontext.TestingTFromContext(ctx); t != nil {
		logger = createTestingLogger(t)
	} else {
		logger = teeLoggers(
			createDefaultLogger(ctx),
			createFileLogger(ctx),
		)
	}
	return logger.Sugar()
}

func teeLoggers(logger1 *zap.Logger, logger2 *zap.Logger) *zap.Logger {
	return zap.New(zapcore.NewTee(
		logger1.Core(),
		logger2.Core(),
	))
}

func createFileLogger(ctx context.Context) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder

	logFile := LogFileFrom(ctx)
	if logFile == nil {
		fatal(errors.New("no log file in context"))
	}
	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		zapcore.AddSync(logFile),
		zapcore.DebugLevel,
	))
}

func createDefaultLogger(ctx context.Context) *zap.Logger {
	prtr := output.PrinterFrom(ctx)
	errout := prtr.ErrOrStderr()
	ec := zap.NewDevelopmentEncoderConfig()
	ec.EncodeLevel = zapcore.CapitalLevelEncoder
	if environment.SupportsColor() {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	ec.EncodeTime = ElapsedMillisTimeEncoder(time.Now())
	ec.ConsoleSeparator = " "

	lvl := activeLogLevel(LogLevelFromContext(ctx))
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(ec),
		zapcore.AddSync(errout),
		lvl,
	))
	if !term.IsFancy(errout) {
		ec = zap.NewProductionEncoderConfig()
		logger = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(ec),
			zapcore.AddSync(errout),
			lvl,
		))
	}
	return logger
}

func createTestingLogger(t pkgcontext.TestingT) *zap.Logger {
	lvl := activeLogLevel(zapcore.DebugLevel)
	return zaptest.NewLogger(t, zaptest.WrapOptions(
		zap.AddCaller(),
	), zaptest.Level(lvl))
}

func activeLogLevel(defaultLevel zapcore.Level) zapcore.Level {
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		l, err := zapcore.ParseLevel(lvl)
		if err != nil {
			fatal(errors.WithStack(err))
		}
		return l
	}
	return defaultLevel
}

// ElapsedMillisTimeEncoder is a time encoder using elapsed time since the
// logger setup.
func ElapsedMillisTimeEncoder(setupTime time.Time) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendInt64(t.Sub(setupTime).Milliseconds())
	}
}
