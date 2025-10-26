package typedquery

import (
	"context"
	"fmt"
	"hermes-relay/internal/lib/utils"
	"hermes-relay/internal/projection"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Input struct {
	Path map[string]string
	Body []byte
}

type Output struct {
	StatusCode int
	Body       []byte
}

type ProcessorFunc func(ctx context.Context, in Input) Output

func ToRoute(processor ProcessorFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, fmt.Sprintf("internal error: %v", rec), http.StatusInternalServerError)
			}
		}()

		pathParams := make(map[string]string)
		rctx := chi.RouteContext(r.Context())
		if rctx != nil {
			for i, key := range rctx.URLParams.Keys {
				if i < len(rctx.URLParams.Values) {
					pathParams[key] = rctx.URLParams.Values[i]
				}
			}
		}

		var body []byte
		if r.ContentLength > 0 {
			var err error
			body, err = io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		input := Input{
			Path: pathParams,
			Body: body,
		}

		output := processor(r.Context(), input)

		w.WriteHeader(output.StatusCode)
		utils.Must(w.Write(output.Body))
	}
}

func Route[T, Q, R any](
	store *projection.Store[T],
	queryFn QueryFunc[T, Q, R],
) http.HandlerFunc {
	return ToRoute(Query[T, Q, R](store, queryFn))
}

func RouteWithMap[T, Q, R, Y any](
	store *projection.Store[T],
	queryFn QueryFunc[T, Q, R],
	mapFn func(R) Y,
) http.HandlerFunc {
	return ToRoute(QueryWithMap[T, Q, R, Y](store, queryFn, mapFn))
}
