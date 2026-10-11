package handler

import (
	"net/http"
)


func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("club-service is active and routing correctly"))
}