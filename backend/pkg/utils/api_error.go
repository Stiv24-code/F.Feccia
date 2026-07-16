package utils

// APIError is an error carrying an HTTP status code.
// Services can return it and handlers can keep using HandleDatabaseError.
type APIError struct {
	Code    int
	Message string
}

func (e APIError) Error() string { return e.Message }
func (e APIError) StatusCode() int {
	return e.Code
}

func NewAPIError(code int, message string) error {
	return APIError{Code: code, Message: message}
}

