package handlers

import (
	"encoding/json"
	"hermes-relay/internal/persistence"
	"hermes-relay/internal/utils"
	"log/slog"
	"net/http"
	"strings"
)

func RESTHandler[T any](store *persistence.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, r.Pattern)

		slog.Debug(
			"REST request",
			"path", path,
			"pattern", r.Pattern)

		if path == "" || path == "/" {
			// GET /things
			items, err := persistence.GetAll[T](store)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			utils.WarnErr(json.NewEncoder(w).Encode(items))
			return
		}

		// GET /things/{id}
		id := strings.TrimPrefix(path, "/")
		item, err := persistence.GetByID[T](store, id)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		utils.WarnErr(json.NewEncoder(w).Encode(item))
	}
}
