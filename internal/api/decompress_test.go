package api

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	th "nabu-storage/internal/lib/testutil"
)

func TestRequestBodyMiddleware(t *testing.T) {
	const maxBytes = 1024

	tests := []struct {
		Name         string
		Encoding     string
		Body         []byte
		ExpectStatus int
		ExpectBody   string
	}{
		{
			Name:         "plain body passes through",
			Body:         []byte("# Notes"),
			ExpectStatus: http.StatusOK,
			ExpectBody:   "# Notes",
		},
		{
			Name:         "gzip body is decompressed",
			Encoding:     "gzip",
			Body:         gzipped(t, "# Notes"),
			ExpectStatus: http.StatusOK,
			ExpectBody:   "# Notes",
		},
		{
			Name:         "encoding token is case insensitive",
			Encoding:     "GZIP",
			Body:         gzipped(t, "# Notes"),
			ExpectStatus: http.StatusOK,
			ExpectBody:   "# Notes",
		},
		{
			Name:         "corrupt gzip is rejected",
			Encoding:     "gzip",
			Body:         []byte("not actually gzip"),
			ExpectStatus: http.StatusBadRequest,
			ExpectBody:   "invalid gzip body",
		},
		{
			Name:         "oversized plain body is rejected",
			Body:         bytes.Repeat([]byte("a"), maxBytes+1),
			ExpectStatus: http.StatusRequestEntityTooLarge,
			ExpectBody:   "request body too large",
		},
		{
			Name:         "gzip bomb is rejected on the decompressed size",
			Encoding:     "gzip",
			Body:         gzipped(t, strings.Repeat("a", maxBytes*64)),
			ExpectStatus: http.StatusRequestEntityTooLarge,
			ExpectBody:   "request body too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/commands/x", bytes.NewReader(tt.Body))
			if tt.Encoding != "" {
				req.Header.Set("Content-Encoding", tt.Encoding)
			}
			rec := httptest.NewRecorder()

			bodyLimitedChain(maxBytes, http.HandlerFunc(echoBody)).ServeHTTP(rec, req)

			th.AssertEqual(t, rec.Code, tt.ExpectStatus, "status")
			if !strings.Contains(rec.Body.String(), tt.ExpectBody) {
				t.Fatalf("body: expected to contain %q, got %q", tt.ExpectBody, rec.Body.String())
			}
		})
	}
}

func TestDecompressGzipRewritesLengthMetadata(t *testing.T) {
	body := gzipped(t, "# Notes")

	var gotLength int64
	var gotEncoding string
	probe := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotLength = r.ContentLength
		gotEncoding = r.Header.Get("Content-Encoding")
	})

	req := httptest.NewRequest(http.MethodPost, "/commands/x", bytes.NewReader(body))
	req.Header.Set("Content-Encoding", "gzip")
	req.ContentLength = int64(len(body))

	DecompressGzip(1024)(probe).ServeHTTP(httptest.NewRecorder(), req)

	th.AssertEqual(t, gotLength, int64(-1), "content length")
	th.AssertEqual(t, gotEncoding, "", "content encoding header")
}

func bodyLimitedChain(maxBytes int64, next http.Handler) http.Handler {
	return LimitRequestBody(maxBytes)(DecompressGzip(maxBytes)(next))
}

func echoBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	_, _ = w.Write(body)
}

func gzipped(t *testing.T, content string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(content)); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}
