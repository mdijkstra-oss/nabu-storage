package typedquery

import (
	"fmt"
	httphandlers "hermes-relay/internal/handlers/http"
	"hermes-relay/internal/lib/utils"
	"hermes-relay/internal/projection"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ProcessorFunc func(in httphandlers.Request) httphandlers.Response

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

		queryParams := make(map[string]string)
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0]
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

		request := httphandlers.Request{
			Path:  pathParams,
			Query: queryParams,
			Body:  body,
		}

		response := processor(request)

		w.WriteHeader(response.StatusCode)
		utils.Must(w.Write(response.Body))
	}
}

func withStoreFromRequest[T projection.Entity](
	getStore func(r *http.Request) *projection.Store[T],
	handler func(*projection.Store[T]) http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := getStore(r)
		if store == nil {
			http.Error(w, "store not available", http.StatusInternalServerError)
			return
		}
		handler(store)(w, r)
	}
}

func QueryRoute[T projection.Entity, Q, R any](
	getStore func(r *http.Request) *projection.Store[T],
	filterFunc projection.FilterFunc[T, Q, R],
) http.HandlerFunc {
	return withStoreFromRequest(getStore, func(store *projection.Store[T]) http.HandlerFunc {
		exec := projection.BindQuery(store, filterFunc)
		return ToRoute(Query[Q, []R](exec))
	})
}

func QueryOneRoute[T projection.Entity, Q, R any](
	getStore func(r *http.Request) *projection.Store[T],
	filterFunc projection.FilterFunc[T, Q, R],
) http.HandlerFunc {
	return withStoreFromRequest(getStore, func(store *projection.Store[T]) http.HandlerFunc {
		exec := projection.BindQueryOne(store, filterFunc)
		return ToRoute(Query[Q, *R](exec))
	})
}
