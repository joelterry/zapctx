// Package zapctx is a wrapper around the standard library package context
// meant to add *zap.Logger functionality to context.Context.
package zapctx

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type ZapLogger interface {
	DPanic(string, ...zap.Field)
	Debug(string, ...zap.Field)
	Error(string, ...zap.Field)
	Fatal(string, ...zap.Field)
	Info(string, ...zap.Field)
	Panic(string, ...zap.Field)
	Warn(string, ...zap.Field)
}

type LoggerContext interface {
	context.Context
	ZapLogger
	With(...zap.Field) LoggerContext
	WithOptions(...zap.Option) LoggerContext
}

var baseLogger *zap.Logger = zap.NewNop()

type loggerKey struct{}

type loggerCtx struct {
	context.Context
	*zap.Logger
}

func (lc *loggerCtx) Value(key interface{}) interface{} {
	if key == (loggerKey{}) {
		return lc
	}
	return lc.Context.Value(key)
}

func (lc *loggerCtx) With(fields ...zap.Field) LoggerContext {
	return &loggerCtx{lc, lc.Logger.With(fields...)}
}

func (lc *loggerCtx) WithOptions(options ...zap.Option) LoggerContext {
	return &loggerCtx{lc, lc.Logger.WithOptions(options...)}
}

func Logger(parent context.Context) LoggerContext {
	if lc, ok := parent.(LoggerContext); ok {
		return lc
	}
	if lc, ok := parent.Value(loggerKey{}).(*loggerCtx); ok {
		return &loggerCtx{parent, lc.Logger}
	}
	return &loggerCtx{parent, baseLogger}
}

func WithLogger(parent context.Context, logger *zap.Logger) LoggerContext {
	return &loggerCtx{parent, logger}
}

func WithFields(parent context.Context, fields ...zap.Field) LoggerContext {
	if lc, ok := parent.Value(loggerKey{}).(*loggerCtx); ok {
		return &loggerCtx{parent, lc.Logger.With(fields...)}
	}
	return &loggerCtx{parent, baseLogger.With(fields...)}
}

func WithOptions(parent context.Context, options ...zap.Option) LoggerContext {
	if lc, ok := parent.Value(loggerKey{}).(*loggerCtx); ok {
		return &loggerCtx{parent, lc.Logger.WithOptions(options...)}
	}
	return &loggerCtx{parent, baseLogger.WithOptions(options...)}
}

// The rest of this file is a wrapper around the standard context package,
// with all functions returning a LoggerContext instead of a context.Context.
// This allows users to use zapctx as a replacement for context.

// The Context alias is for users of this package.
// LoggerContext is used within the package to avoid confusion.
type Context = LoggerContext

type CancelFunc = context.CancelFunc

var Canceled = context.Canceled
var DeadlineExceeded = context.DeadlineExceeded

func Background() LoggerContext {
	return Logger(context.Background())
}

func TODO() LoggerContext {
	return Logger(context.TODO())
}

func WithCancel(parent context.Context) (LoggerContext, CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	return Logger(ctx), cancel
}

func WithDeadline(parent context.Context, d time.Time) (LoggerContext, CancelFunc) {
	ctx, cancel := context.WithDeadline(parent, d)
	return Logger(ctx), cancel
}

func WithTimeout(parent context.Context, timeout time.Duration) (LoggerContext, CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	return Logger(ctx), cancel
}

func WithValue(parent context.Context, key, val interface{}) LoggerContext {
	ctx := context.WithValue(parent, key, val)
	return Logger(ctx)
}
