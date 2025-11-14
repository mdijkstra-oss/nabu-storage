package http

import (
	"hermes-relay/internal/cqrs/dispatch"
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
		WriteResponse(w, response)
	}
}
