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

func httpHandler(processor func(context.Context, Input, cqrs.PublishFunc) Output, publish cqrs.PublishFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		output := processor(r.Context(), Input{Body: body}, publish)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(output.StatusCode)
		utils.Must(w.Write(output.Body))
	}
}
