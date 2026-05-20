package utils

import "net/http"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e AppError) Error() string {
	return e.Message
}

var (
	ErrUnauthorized = AppError{
		Code:    "unauthorized",
		Message: "Unauthorized",
		Status:  http.StatusUnauthorized,
	}
	ErrForbidden = AppError{
		Code:    "forbidden",
		Message: "Forbidden",
		Status:  http.StatusForbidden,
	}
	ErrNotFound = AppError{
		Code:    "not_found",
		Message: "Not found",
		Status:  http.StatusNotFound,
	}
	ErrConflict = AppError{
		Code:    "conflict",
		Message: "Conflict",
		Status:  http.StatusConflict,
	}
	ErrInvalidInput = AppError{
		Code:    "invalid_input",
		Message: "Invalid input",
		Status:  http.StatusBadRequest,
	}
)

func NewAppError(code, message string, status int) AppError {
	return AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}
