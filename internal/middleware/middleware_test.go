package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDGenerated(t *testing.T) {
	var capturedID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	respID := rec.Header().Get(RequestIDHeader)
	if respID == "" {
		t.Fatal("expected X-Request-ID header in response")
	}
	if capturedID != respID {
		t.Fatalf("expected context ID %q to match response header %q", capturedID, respID)
	}
	if len(respID) != 32 {
		t.Fatalf("expected 32 hex char request ID, got length %d: %q", len(respID), respID)
	}
}

func TestRequestIDPropagated(t *testing.T) {
	incomingID := "custom-req-id-12345"
	var capturedID string

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, incomingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	respID := rec.Header().Get(RequestIDHeader)
	if respID != incomingID {
		t.Fatalf("expected response header %q, got %q", incomingID, respID)
	}
	if capturedID != incomingID {
		t.Fatalf("expected context ID %q, got %q", incomingID, capturedID)
	}
}

func TestGetRequestIDEmptyContext(t *testing.T) {
	id := GetRequestID(context.Background())
	if id != "" {
		t.Fatalf("expected empty string for background context, got %q", id)
	}
}

func TestLoggerMiddleware(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok response"))
	})

	mw := Logger(logger)
	chain := RequestID(mw(nextHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/urls", nil)
	rec := httptest.NewRecorder()

	chain.ServeHTTP(rec, req)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, `"msg":"http request handled"`) {
		t.Fatalf("expected log msg 'http request handled', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"path":"/api/v1/urls"`) {
		t.Fatalf("expected path in log output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"status":200`) {
		t.Fatalf("expected status 200 in log output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"request_id"`) {
		t.Fatalf("expected request_id in log output, got: %s", logOutput)
	}
}

func TestLoggerMiddlewareWarnAndErrorLevels(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	// Test 404 -> WARN
	mw := Logger(logger)
	notFoundHandler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	notFoundHandler.ServeHTTP(rec, req)

	if !strings.Contains(logBuf.String(), `"level":"WARN"`) {
		t.Fatalf("expected WARN level for 404, got: %s", logBuf.String())
	}

	logBuf.Reset()

	// Test 500 -> ERROR
	serverErrorHandler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/error", nil)
	serverErrorHandler.ServeHTTP(rec, req)

	if !strings.Contains(logBuf.String(), `"level":"ERROR"`) {
		t.Fatalf("expected ERROR level for 500, got: %s", logBuf.String())
	}
}

func TestRecovererMiddleware(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected nil pointer")
	})

	mw := Recoverer(logger)
	chain := RequestID(mw(panicHandler))

	req := httptest.NewRequest(http.MethodGet, "/panic-route", nil)
	rec := httptest.NewRecorder()

	// Ensure the test itself doesn't crash
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 on panic recovery, got %d", rec.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, `"panic recovered in http handler"`) {
		t.Fatalf("expected panic message in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `unexpected nil pointer`) {
		t.Fatalf("expected panic detail in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"stack"`) {
		t.Fatalf("expected stack trace in logs, got: %s", logOutput)
	}
}
