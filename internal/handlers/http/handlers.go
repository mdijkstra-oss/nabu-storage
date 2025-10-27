package http

import (
	"context"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/lib/utils"
	"io"
	"net/http"
)

func CommandHandler(publish cqrs.PublishFunc) http.HandlerFunc {
	return httpHandler(ProcessCommand, publish)
}

//func EventHandler(publish cqrs.PublishFunc) http.HandlerFunc {
//	return httpHandler(ProcessEvent, publish)
//}

func httpHandler(processor func(context.Context, Request, cqrs.PublishFunc) Response, publish cqrs.PublishFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := processor(r.Context(), Request{Body: body}, publish)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.StatusCode)
		utils.Must(w.Write(response.Body))
	}
}
