package slash_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"htdvisser.dev/slash"
)

func buildSlashRequest(signingSecret string, time time.Time, params url.Values) *http.Request {
	body := params.Encode()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	timestamp := strconv.FormatInt(time.Unix(), 10)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	hash := hmac.New(sha256.New, []byte(signingSecret))
	fmt.Fprint(hash, "v0:", timestamp, ":")
	hash.Write([]byte(body))
	sum := hash.Sum(nil)
	req.Header.Set("X-Slack-Signature", fmt.Sprintf("v0=%x", sum))
	return req
}

func TestRouter(t *testing.T) {
	t.Run("NoGET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		router := slash.NewRouter("")
		router.ServeHTTP(rec, req)
		if code := rec.Result().StatusCode; code != http.StatusMethodNotAllowed {
			t.Fatalf("Expected GET request to be rejected, got %v", code)
		}
	})

	t.Run("NoJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router := slash.NewRouter("")
		router.ServeHTTP(rec, req)
		if code := rec.Result().StatusCode; code != http.StatusBadRequest {
			t.Fatalf("Expected JSON request to be rejected, got %v", code)
		}
	})

	t.Run("PastRequest", func(t *testing.T) {
		req := buildSlashRequest("00000000000000000000000000000000", time.Now().Add(-10*time.Minute), nil)
		rec := httptest.NewRecorder()
		router := slash.NewRouter("00000000000000000000000000000000")
		router.ServeHTTP(rec, req)
		if code := rec.Result().StatusCode; code != http.StatusUnauthorized {
			t.Fatalf("Expected past request to be rejected, got %v", code)
		}
	})

	t.Run("FutureRequest", func(t *testing.T) {
		req := buildSlashRequest("00000000000000000000000000000000", time.Now().Add(10*time.Minute), nil)
		rec := httptest.NewRecorder()
		router := slash.NewRouter("00000000000000000000000000000000")
		router.ServeHTTP(rec, req)
		if code := rec.Result().StatusCode; code != http.StatusUnauthorized {
			t.Fatalf("Expected future request to be rejected, got %v", code)
		}
	})

	t.Run("KeyMismatch", func(t *testing.T) {
		req := buildSlashRequest("11111111111111111111111111111111", time.Now(), nil)
		rec := httptest.NewRecorder()
		router := slash.NewRouter("00000000000000000000000000000000")
		router.ServeHTTP(rec, req)
		if code := rec.Result().StatusCode; code != http.StatusUnauthorized {
			t.Fatalf("Expected old request to be rejected, got %v", code)
		}
	})

	params := url.Values{}
	params.Set("command", "/ping")
	params.Set("user_name", "user")
	params.Set("team_domain", "example")

	ping := func(router *slash.Router) *http.Response {
		req := buildSlashRequest("00000000000000000000000000000000", time.Now(), params)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec.Result()
	}

	t.Run("UnknownCommand", func(t *testing.T) {
		res := ping(slash.NewRouter("00000000000000000000000000000000"))
		if code := res.StatusCode; code != http.StatusOK {
			t.Fatalf("Expected request to be accepted, got %v", code)
		}
		var m map[string]interface{}
		json.NewDecoder(res.Body).Decode(&m)
		if text, ok := m["text"].(string); !ok || !strings.HasPrefix(text, "Sorry") {
			t.Fatalf("Expected response to contain text \"Sorry\", got %v", m["text"])
		}
	})

	t.Run("WithResponse", func(t *testing.T) {
		router := slash.NewRouter("00000000000000000000000000000000")
		router.Register("/ping", func(ctx context.Context, req slash.Request) interface{} {
			return map[string]interface{}{
				"text": "pong",
			}
		})
		res := ping(router)
		if code := res.StatusCode; code != http.StatusOK {
			t.Fatalf("Expected request to be accepted, got %v", code)
		}
		var m map[string]interface{}
		json.NewDecoder(res.Body).Decode(&m)
		if text, ok := m["text"].(string); !ok || text != "pong" {
			t.Fatalf("Expected response to contain text \"pong\", got %v", m["text"])
		}
	})

	t.Run("WithoutResponse", func(t *testing.T) {
		router := slash.NewRouter("00000000000000000000000000000000")
		router.Register("/ping", func(ctx context.Context, req slash.Request) interface{} {
			return nil
		})
		res := ping(router)
		if code := res.StatusCode; code != http.StatusOK {
			t.Fatalf("Expected request to be accepted, got %v", code)
		}
	})

	t.Run("WithPanic", func(t *testing.T) {
		router := slash.NewRouter("00000000000000000000000000000000")
		router.Register("/ping", func(ctx context.Context, req slash.Request) interface{} {
			panic("no")
		})
		res := ping(router)
		if code := res.StatusCode; code != http.StatusOK {
			t.Fatalf("Expected request to be accepted, got %v", code)
		}
		var m map[string]interface{}
		json.NewDecoder(res.Body).Decode(&m)
		if text, ok := m["text"].(string); !ok || !strings.HasPrefix(text, "Sorry") {
			t.Fatalf("Expected response to contain text \"Sorry\", got %v", m["text"])
		}
	})
}
