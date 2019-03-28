package slash

import (
	"context"
)

// Logger is the logger interface used by the slash command Router.
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

type logContextKeyType struct{}

var logContextKey logContextKeyType

func newContextWithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, logContextKey, logger)
}

type noopLogger struct{}

func (noopLogger) Print(_ ...interface{})            {}
func (noopLogger) Printf(_ string, _ ...interface{}) {}

func loggerFromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(logContextKey).(Logger); ok {
		return logger
	}
	return noopLogger{}
}
