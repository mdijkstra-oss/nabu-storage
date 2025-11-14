package http

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/utils"
	"net/http"
)

var validate = validator.New()

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

	if err := validate.Struct(query); err != nil {
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

func ProjectQuery[Q, R any](
	registryState *registry.RegistryState,
	handler func(Q, project.Project) R,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		if projectID == "" {
			respondWithError(w, &utils.ValidationError{Message: "projectId is required"})
			return
		}

		proj := registryState.GetProject(projectID)
		if proj == nil {
			respondWithError(w, &utils.NotFoundError{Message: "project not found"})
			return
		}

		if !proj.IsHealthy() {
			respondWithError(w, errors.New("project is in unhealthy state due to corrupted data"))
			return
		}

		query, err := ParseQuery[Q](r)
		if err != nil {
			respondWithError(w, err)
			return
		}

		result := handler(query, *proj)
		respondWithJSON(w, result)
	}
}

func respondWithError(w http.ResponseWriter, err error) {
	response := typedErrorOutput(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	utils.Should(w.Write(response.Body))
}

func respondWithJSON(w http.ResponseWriter, result any) {
	if utils.IsNilPtr(result) {
		respondWithError(w, &utils.NotFoundError{Message: "not found"})
		return
	}

	response := successQueryOutput(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	utils.Should(w.Write(response.Body))
}
