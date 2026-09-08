package utils

import (
	"encoding/json"
	"net/http"
)

type JSONErrorResponse struct {
	Error     string `json:"error"`
	IsSuccess bool   `json:"isSuccess"`
}

func RespondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := JSONErrorResponse{Error: message, IsSuccess: false}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
	}
}
