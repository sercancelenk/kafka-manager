package http

import "net/http"

type Error struct {
	Code    int    `json:"code"`
	Err     string `json:"error"`
	Message string `json:"message"`
}

func (h Error) Error() string {
	return h.Message
}

func New(status int, message string) *Error {
	return &Error{
		Code:    status,
		Err:     http.StatusText(status),
		Message: message,
	}
}
