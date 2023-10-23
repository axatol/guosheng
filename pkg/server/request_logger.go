package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	_ middleware.LogFormatter = (*requestLogFormatter)(nil)
	_ middleware.LogEntry     = (*requestLogEntry)(nil)
)

type requestLogFormatter struct{}

func (lf requestLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	uri := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	log := log.With().
		Str("request_id", middleware.GetReqID(r.Context())).
		Str("method", r.Method).
		Str("proto", r.Proto).
		Str("uri", uri).
		Logger()

	return requestLogEntry{log}
}

type requestLogEntry struct{ log zerolog.Logger }

func (le requestLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	le.log.Debug().
		Int("status", status).
		Int("bytes_written", bytes).
		Str("elapsed", elapsed.String()).
		Any("metadata", extra).
		Send()
}

func (le requestLogEntry) Panic(v any, stack []byte) {
	le.log.Debug().
		Any("recovery", v).
		Bytes("recovery_stack", stack).
		Stack().
		Send()
}
