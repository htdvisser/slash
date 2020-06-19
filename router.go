package slash

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// HandlerFunc is a function that handles a slash command.
type HandlerFunc func(ctx context.Context, req Request) interface{}

// MiddlewareFunc is a function that can be used as middleware in slash command handling.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// Router routes slash commands to their handlers.
type Router struct {
	getLogger             func(r *http.Request) Logger
	getTimeDifference     func(timestamp int64) (time.Duration, bool)
	signingSecret         string
	middleware            MiddlewareFunc
	commands              map[string]HandlerFunc
	commandUnknownHandler HandlerFunc
	commandFailedHandler  HandlerFunc
}

// NewRouter creates a new router for slash commands.
func NewRouter(signingSecret string, opts ...RouterOption) *Router {
	r := &Router{
		getLogger:             getLogger,
		getTimeDifference:     getTimeDifference,
		signingSecret:         signingSecret,
		commands:              make(map[string]HandlerFunc),
		commandUnknownHandler: commandUnknownHandler,
		commandFailedHandler:  commandFailedHandler,
	}
	for _, opt := range opts {
		opt.apply(r)
	}
	return r
}

// RegisterCommand registers a slash command and its handler to the Router.
func (sr *Router) RegisterCommand(command string, handler HandlerFunc) {
	sr.commands[command] = handler
}

const maxBodySize = 1024 * 1024

func (sr *Router) verifyRequest(header http.Header, body []byte) error {
	if sr.signingSecret == "" {
		return nil
	}
	timestampHeader := header.Get("X-Slack-Request-Timestamp")
	timestamp, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid request timestamp: %v", err)
	}
	if difference, ok := sr.getTimeDifference(timestamp); !ok {
		return fmt.Errorf("invalid request timestamp: difference of %s too big", difference)
	}
	signatureHeader := header.Get("X-Slack-Signature")
	if !strings.HasPrefix(signatureHeader, "v0=") {
		return fmt.Errorf("invalid request signature: does not contain version")
	}
	signature, err := hex.DecodeString(strings.TrimPrefix(signatureHeader, "v0="))
	if err != nil {
		return fmt.Errorf("invalid request signature: %v", err)
	}
	hash := hmac.New(sha256.New, []byte(sr.signingSecret))
	if _, err = fmt.Fprint(hash, "v0:", timestampHeader, ":"); err != nil {
		return err
	}
	if _, err = hash.Write(body); err != nil {
		return err
	}
	if !hmac.Equal(hash.Sum(nil), signature) {
		return errors.New("invalid request signature")
	}
	return nil
}

// HandleCommand reads the command from the request header+body and returns a response.
func (sr *Router) HandleCommand(ctx context.Context, header http.Header, body []byte) (s int, h http.Header, b []byte, err error) {
	if contentType := header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
		return http.StatusBadRequest, plainResponseHeader, []byte(fmt.Sprintf("requires application/x-www-form-urlencoded, not %s", contentType)), nil
	}
	if err := sr.verifyRequest(header, body); err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, err)
	}
	params, err := url.ParseQuery(string(body))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid body: %v", err))
	}
	req := Request(params)
	handler, ok := sr.commands[req.Command()]
	if !ok {
		return jsonResponse(ctx, sr.commandUnknownHandler(ctx, req))
	}
	logger := loggerFromContext(ctx)
	logger.Printf("handling command `%s` for @%s of team %s", req.Command(), req.UserName(), req.TeamDomain())
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic in command handler: %v\n%s", r, string(debug.Stack()))
			s, h, b, err = jsonResponse(ctx, sr.commandFailedHandler(ctx, req))
		}
	}()
	if sr.middleware != nil {
		handler = sr.middleware(handler)
	}
	return jsonResponse(ctx, handler(ctx, req))
}

// ServeHTTP implements http.Handler.
func (sr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ctx = newContextWithLogger(ctx, sr.getLogger(r))

	if r.Method != http.MethodPost {
		s, h, b, _ := errorResponse(ctx, http.StatusMethodNotAllowed, fmt.Errorf("requires POST, not %s", r.Method))
		writeResponse(w, s, h, b)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		s, h, b, _ := errorResponse(ctx, http.StatusBadRequest, err)
		writeResponse(w, s, h, b)
		return
	}

	s, h, b, err := sr.HandleCommand(ctx, r.Header, body)
	if err != nil {
		s, h, b, _ = errorResponse(ctx, http.StatusInternalServerError, err)
	}
	writeResponse(w, s, h, b)
}
