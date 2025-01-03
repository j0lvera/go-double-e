package server

// ErrorResponse is the response sent back to the client when an error occurs
type ErrorResponse struct {
	Status  int         `json:"status"`
	Message interface{} `json:"message,omitempty"`
}

// General error message to return to the client
const (
	ErrInternalServerError = "Internal Server Error"
	ErrInvalidRequest      = "Invalid request"
	ErrUnauthorized        = "Unauthorized"
	ErrForbidden           = "Forbidden"
	ErrNotFound            = "Not Found"

	ErrUserAlreadyExists  = "Email already registered"
	ErrInvalidCredentials = "Invalid credentials"
)
