package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kimmykuang/selfskill/internal/service"
)

// httpError writes a JSON error response with the given status.
func httpError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// httpFromErr maps a service-layer error to an HTTP status. Sentinel errors
// in service map to 4xx; anything else is 500.
func httpFromErr(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, service.ErrNotFound):
		httpError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAlreadyExists):
		httpError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrConflict):
		httpError(w, http.StatusConflict, err.Error())
	default:
		httpError(w, http.StatusInternalServerError, err.Error())
	}
}
