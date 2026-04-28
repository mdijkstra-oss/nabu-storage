package http

import (
	"compress/gzip"
	"net/http"
)

func DecompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid gzip body")
			return
		}
		defer gz.Close()

		r.Body = gz
		r.Header.Del("Content-Encoding")
		next.ServeHTTP(w, r)
	})
}
