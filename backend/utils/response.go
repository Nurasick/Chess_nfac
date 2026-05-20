package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Data    interface{} `json:"data"`
	Error   *string     `json:"error"`
	Message string      `json:"message"`
}

type PaginatedResponse struct {
	Data    interface{} `json:"data"`
	Error   *string     `json:"error"`
	Message string      `json:"message"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Data:    data,
		Error:   nil,
		Message: "success",
	})
}

func RespondError(w http.ResponseWriter, status int, code, message string) {
	RespondAppError(w, NewAppError(code, message, status))
}

func RespondAppError(w http.ResponseWriter, appErr AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)
	json.NewEncoder(w).Encode(APIResponse{
		Data:    nil,
		Error:   &appErr.Code,
		Message: appErr.Message,
	})
}

func RespondPaginated(w http.ResponseWriter, status int, data interface{}, total, page, limit int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(PaginatedResponse{
		Data:    data,
		Error:   nil,
		Message: "success",
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}
