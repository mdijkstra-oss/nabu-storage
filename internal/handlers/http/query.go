package http

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"net/http"
)

func ParseQuery[Q any](r *http.Request) (Q, error) {
	var query Q

	if err := bindPathParams(r, &query); err != nil {
		return query, &utils.ValidationError{Message: err.Error()}
	}

	if err := bindQueryParams(r, &query); err != nil {
		return query, &utils.ValidationError{Message: err.Error()}
	}

	if err := utils.ApplyDefaults(&query); err != nil {
		return query, err
	}

	if err := utils.Validate.Struct(query); err != nil {
		return query, utils.ToValidationError(err)
	}

	return query, nil
}

func bindPathParams(r *http.Request, dst any) error {
	pathParams := make(map[string]string)
	rctx := chi.RouteContext(r.Context())
	if rctx != nil {
		for i, key := range rctx.URLParams.Keys {
			if i < len(rctx.URLParams.Values) {
				pathParams[key] = rctx.URLParams.Values[i]
			}
		}
	}
	return utils.BindParams(pathParams, dst, "path")
}

func bindQueryParams(r *http.Request, dst any) error {
	queryParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}
	return utils.BindParams(queryParams, dst, "query")
}

func Query[Q, R any](handler func(Q) R) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := ParseQuery[Q](r)
		if err != nil {
			respondWithError(w, err)
			return
		}

		result := handler(query)
		respondWithJSON(w, result)
	}
}

func StoreQuery[Q, R any](
	store *registry.Store,
	handler func(Q, string, *registry.Store) R,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		if projectID == "" {
			respondWithError(w, &utils.ValidationError{Message: "projectId is required"})
			return
		}

		proj := projection.Read(store, func(reg *registry.Registry) *registry.Registry {
			return reg
		})
		if registry.GetProject(proj, projectID) == nil {
			respondWithError(w, &utils.NotFoundError{Message: "project not found"})
			return
		}

		query, err := ParseQuery[Q](r)
		if err != nil {
			respondWithError(w, err)
			return
		}

		result := handler(query, projectID, store)
		respondWithJSON(w, result)
	}
}

func respondWithError(w http.ResponseWriter, err error) {
	response := typedErrorOutput(err)
	slog.Error("query error", "status", response.StatusCode, "response", string(response.Body))
	WriteResponse(w, response)
}

func respondWithJSON(w http.ResponseWriter, result any) {
	if utils.IsNilPtr(result) {
		response := typedErrorOutput(&utils.NotFoundError{Message: "not found"})
		slog.Error("query error", "status", response.StatusCode, "response", string(response.Body))
		WriteResponse(w, response)
		return
	}
	WriteResponse(w, successQueryOutput(result))
}
