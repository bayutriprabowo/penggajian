package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/bayutriprabowo/penggajian/internal/dto"
	"github.com/bayutriprabowo/penggajian/internal/service"
)

const maxBodyBytes = 1 << 20

func writeJSON(w http.ResponseWriter, status int, resp dto.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeOK(w http.ResponseWriter, message string, data any) {
	writeJSON(w, http.StatusOK, dto.Success(message, data))
}

func writeCreated(w http.ResponseWriter, message string, data any) {
	writeJSON(w, http.StatusCreated, dto.Success(message, data))
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrBadRequest):
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
	case errors.Is(err, service.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, dto.Error(err.Error()))
	case errors.Is(err, service.ErrForbidden):
		writeJSON(w, http.StatusForbidden, dto.Error(err.Error()))
	case errors.Is(err, service.ErrNotFound):
		writeJSON(w, http.StatusNotFound, dto.Error(err.Error()))
	case errors.Is(err, service.ErrConflict):
		writeJSON(w, http.StatusConflict, dto.Error(err.Error()))
	default:
		log.Printf("[error] %v", err)
		writeJSON(w, http.StatusInternalServerError, dto.Error("terjadi kesalahan server"))
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("body harus berisi satu objek JSON")
	}
	return nil
}

func pathID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, service.ErrBadRequest
	}
	return uint(id), nil
}

func queryUint(r *http.Request, key string, def uint) uint {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return def
	}
	return uint(n)
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func queryFloat(r *http.Request, key string, def float64) float64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
}
