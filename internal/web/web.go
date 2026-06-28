package web

import (
	"encoding/json"
	"io"
	"net/http"
)

const MaxBodySize = 1048576 // 1MB

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

func Decode(r *http.Request, v any) error {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func ReadBody(r *http.Request) ([]byte, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodySize)
	return io.ReadAll(r.Body)
}
