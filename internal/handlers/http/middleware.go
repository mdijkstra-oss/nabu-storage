package http

import (
	"hermes-relay/internal/lib/normalizer"
	"net/http"
	"reflect"
)

type Headers struct {
	ContentType string
}

var DefaultHeaders = Headers{
	ContentType: "application/json",
}

var MarkDownHeaders = Headers{
	ContentType: "text/markdown; charset=utf-8",
}

func WithHeaders(h any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := reflect.ValueOf(h)
			t := v.Type()

			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				if field.String() == "" {
					continue
				}

				name := normalizer.Kebab(t.Field(i).Name)
				w.Header().Set(name, field.String())
			}

			next.ServeHTTP(w, r)
		})
	}
}
