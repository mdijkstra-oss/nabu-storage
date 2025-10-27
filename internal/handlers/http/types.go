package http

import "hermes-relay/internal/cqrs"

type Request struct {
	Path map[string]string
	Body []byte
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
	Index   int              `json:"index"`
	Success bool             `json:"success"`
	Result  *cqrs.AnyMessage `json:"result,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type batchResponse struct {
	Results      []batchResult `json:"results"`
	Total        int           `json:"total"`
	SuccessCount int           `json:"success_count"`
	FailureCount int           `json:"failure_count"`
}
