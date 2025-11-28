package http

import (
	"hermes-relay/internal/cqrs/commands"
)

type Request struct {
	Path  map[string]string
	Query map[string]string
	Body  []byte
}

type Response struct {
	StatusCode int
	Body       []byte
}

type ErrorResponse struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type batchResult struct {
	Index   int                  `json:"index"`
	Success bool                 `json:"success"`
	Result  *commands.AnyMessage `json:"result,omitempty"`
	Error   string               `json:"error,omitempty"`
}

type batchResponse struct {
	Results      []batchResult `json:"results"`
	Total        int           `json:"total"`
	SuccessCount int           `json:"success_count"`
	FailureCount int           `json:"failure_count"`
}

type partialSuccessResponse struct {
	Success  *commands.AnyMessage `json:"success"`
	Failures map[int]string       `json:"failures"`
}
