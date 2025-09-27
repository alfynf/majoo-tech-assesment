package handler

import "github.com/microcosm-cc/bluemonday"

var (
	// StrictUGC sanitizer policy allows only the most basic text formatting.
	// It's a good default for user-generated content like comments.
	StrictUGC = bluemonday.UGCPolicy()
)

// ErrorResponse defines the standard error response format.
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// LoginResponse defines the successful login response format.
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
