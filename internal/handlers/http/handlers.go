package http

import (
	"encoding/json"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/lib/utils"
	"io"
	"net/http"
)

func CommandHandler(publish dispatch.PublishFunc) http.HandlerFunc {
	return httpHandler(ProcessCommand, publish)
}

//func EventHandler(publish cqrs.PublishFunc) http.HandlerFunc {
//	return httpHandler(ProcessEvent, publish)
//}

func httpHandler(processor func(Request, dispatch.PublishFunc) Response, publish dispatch.PublishFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := processor(Request{Body: body}, publish)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.StatusCode)
		utils.Should(w.Write(response.Body))
	}
}

func ToJSON[T any](handler func(*http.Request) (T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := handler(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
