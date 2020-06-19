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

var plainResponseHeader = func() http.Header {
	header := make(http.Header)
	header.Set("Content-Type", "text/plain")
	return header
}()

func errorResponse(ctx context.Context, status int, err error) (int, http.Header, []byte, error) {
	loggerFromContext(ctx).Printf("fail with error: %v", err)
	return status, plainResponseHeader, []byte(fmt.Sprintf("%s: %v", http.StatusText(status), err)), nil
}

var jsonResponseHeader = func() http.Header {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	return header
}()

func jsonResponse(ctx context.Context, msg interface{}) (int, http.Header, []byte, error) {
	logger := loggerFromContext(ctx)
	if msg == nil {
		logger.Print("finish without response")
		return http.StatusOK, nil, nil, nil
	}
	logger.Print("finish with response")
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	body, err := json.Marshal(msg)
	if err != nil {
		return http.StatusInternalServerError, nil, nil, err
	}
	return http.StatusOK, header, body, nil
}

func writeResponse(w http.ResponseWriter, status int, header http.Header, body []byte) {
	for k := range header {
		w.Header().Set(k, header.Get(k))
	}
	w.WriteHeader(status)
	w.Write(body)
}
