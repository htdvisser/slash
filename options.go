package slash

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// RouterOption sets options on the Router.
type RouterOption interface {
	apply(*Router)
}

type routerOption func(*Router)

func (o routerOption) apply(r *Router) { o(r) }

func getLogger(_ *http.Request) Logger { return log.New(os.Stderr, "slash: ", log.LstdFlags) }

// WithoutLogger returns a RouterOption that disables logging.
func WithoutLogger() RouterOption { return WithLogger(noopLogger{}) }

// WithLogger returns a RouterOption that sets the logger.
func WithLogger(logger Logger) RouterOption {
	return routerOption(func(r *Router) {
		r.getLogger = func(_ *http.Request) Logger { return logger }
	})
}

func getTimeDifference(timestamp int64) (difference time.Duration, ok bool) {
	difference = time.Since(time.Unix(timestamp, 0)).Round(time.Second)
	if difference < 0 {
		difference *= -1
	}
	return difference, difference < 5*time.Minute
}

func commandUnknownHandler(ctx context.Context, req Request) interface{} {
	loggerFromContext(ctx).Printf("unknown command `%s`", req.Command())
	return textResponse(fmt.Sprintf("Sorry %s, I don't know how to handle the `%s` command.", req.UserName(), req.Command()))
}

// WithCommandUnknownHandler returns a RouterOption that sets the handler for unknown commands.
func WithCommandUnknownHandler(f HandlerFunc) RouterOption {
	return routerOption(func(r *Router) {
		r.commandUnknownHandler = f
	})
}
