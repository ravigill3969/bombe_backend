package utils

import (
	"encoding/json"
	"net/http"
)

type JSONSuccessResponse struct {
	Message   string `json:"message"`
	IsSuccess bool   `json:"isSuccess"`
	Data      any    `json:"data,omitempty"`
}

func RespondWithSuccess(w http.ResponseWriter, message string, statusCode int, data any) {
	response := JSONSuccessResponse{
		Message:   message,
		IsSuccess: true,
		Data:      data,
	}

	payload, err := json.Marshal(response)
	if err != nil {
		http.Error(w, `{"message":"Internal server error", "isSuccess":false}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(payload)
}