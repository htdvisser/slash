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
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// HandlerFunc is a function that handles a slash command.
type HandlerFunc func(ctx context.Context, req Request) interface{}

// Router routes slash commands to their handlers.
type Router struct {
	getLogger             func(r *http.Request) Logger
	getTimeDifference     func(timestamp int64) (time.Duration, bool)
	signingSecret         string
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

// Register registers a slash command and its handler to the Router.
func (sr *Router) Register(command string, handler HandlerFunc) {
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

func (sr *Router) handleCommand(ctx context.Context, w http.ResponseWriter, req Request) {
	defer func() {
		if r := recover(); r != nil {
			loggerFromContext(ctx).Printf("panic in command handler: %v\n%s", r, string(debug.Stack()))
			respond(ctx, w, sr.commandFailedHandler(ctx, req))
		}
	}()
	commandName := req.Command()
	if handler, ok := sr.commands[commandName]; ok {
		respond(ctx, w, handler(ctx, req))
		return
	}
	respond(ctx, w, sr.commandUnknownHandler(ctx, req))
}

// ServeHTTP implements http.Handler.
func (sr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ctx = newContextWithLogger(ctx, sr.getLogger(r))

	if r.Method != http.MethodPost {
		respondWithError(ctx, w, http.StatusMethodNotAllowed, fmt.Errorf("requires POST, not %s", r.Method))
		return
	}
	if contentType := r.Header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
		respondWithError(ctx, w, http.StatusBadRequest, fmt.Errorf("requires application/x-www-form-urlencoded, not %s", contentType))
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		respondWithError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := sr.verifyRequest(r.Header, body); err != nil {
		respondWithError(ctx, w, http.StatusUnauthorized, err)
		return
	}

	params, err := url.ParseQuery(string(body))
	if err != nil {
		respondWithError(ctx, w, http.StatusBadRequest, fmt.Errorf("invalid body: %v", err))
		return
	}

	sr.handleCommand(ctx, w, Request(params))
}
