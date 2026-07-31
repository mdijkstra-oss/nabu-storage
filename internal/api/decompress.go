package api

import (
	"compress/gzip"
	"net/http"
	"strings"

	"nabu-storage/internal/lib/utils"
)

const (
	maxRequestBytes      = 8 << 20
	maxDecompressedBytes = 32 << 20
)

func LimitRequestBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func DecompressGzip(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isGzipEncoded(r.Header.Get("Content-Encoding")) {
				next.ServeHTTP(w, r)
				return
			}

			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid gzip body")
				return
			}
			defer func() { utils.ShouldWork(gz.Close()) }()

			r.Body = http.MaxBytesReader(w, gz, maxBytes)
			r.ContentLength = -1
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
			next.ServeHTTP(w, r)
		})
	}
}

func isGzipEncoded(encoding string) bool {
	return strings.EqualFold(strings.TrimSpace(encoding), "gzip")
}
