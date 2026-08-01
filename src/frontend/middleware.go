package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ctxKeyRequestID is the context key for the per-request id.
type ctxKeyRequestID struct{}

// logHandler logs one line per request with method, path, status, bytes and
// duration (upstream uses logrus; this replica uses the stdlib log package
// like the other Go services in the repo).
type logHandler struct {
	next http.Handler
}

type responseRecorder struct {
	b      int
	status int
	w      http.ResponseWriter
}

func (r *responseRecorder) Header() http.Header { return r.w.Header() }

func (r *responseRecorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.w.Write(p)
	r.b += n
	return n, err
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.w.WriteHeader(statusCode)
}

// Flush forwards the flush to the underlying writer. Required by the gRPC
// server when it is served through this recorder (grpc-go needs an
// http.Flusher to write HTTP/2 frames over h2c).
func (r *responseRecorder) Flush() {
	if f, ok := r.w.(http.Flusher); ok {
		f.Flush()
	}
}

func (lh *logHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	start := time.Now()
	rr := &responseRecorder{w: w}

	ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, requestID)
	r = r.WithContext(ctx)

	defer func() {
		log.Printf("http %s %s status=%d bytes=%d took_ms=%d request_id=%s session=%q",
			r.Method, r.URL.Path, rr.status, rr.b, time.Since(start)/time.Millisecond,
			requestID, sessionID(r))
	}()

	lh.next.ServeHTTP(rr, r)
}

// ensureSessionID assigns a persistent session id cookie so the storefront
// can keep a cart per browser session (same behavior as upstream).
func ensureSessionID(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sessionID string
		c, err := r.Cookie(cookieSessionID)
		if err == http.ErrNoCookie {
			sessionID = uuid.New().String()
			http.SetCookie(w, &http.Cookie{
				Name:   cookieSessionID,
				Value:  sessionID,
				MaxAge: cookieMaxAge,
			})
		} else if err != nil {
			return
		} else {
			sessionID = c.Value
		}
		ctx := context.WithValue(r.Context(), ctxKeySessionID{}, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// sessionID returns the current request's session id (assigned by
// ensureSessionID), or "" when absent.
func sessionID(r *http.Request) string {
	if v := r.Context().Value(ctxKeySessionID{}); v != nil {
		return v.(string)
	}
	return ""
}
