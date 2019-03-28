package slash

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type textResponse string

func (r textResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"response_type": "ephemeral",
		"text":          string(r),
	})
}

func respond(ctx context.Context, w http.ResponseWriter, msg interface{}) error {
	logger := loggerFromContext(ctx)
	if msg == nil {
		logger.Print("finish without response")
		return nil
	}
	logger.Print("finish with response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func respondWithError(ctx context.Context, w http.ResponseWriter, code int, err error) error {
	loggerFromContext(ctx).Printf("fail with error: %v", err)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	_, writeErr := fmt.Fprintf(w, "%s: %v", http.StatusText(code), err)
	return writeErr
}
