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

// writeJSON menulis respons JSON dengan status code tertentu.
func writeJSON(w http.ResponseWriter, status int, resp dto.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeOK menulis respons sukses dengan status 200.
func writeOK(w http.ResponseWriter, message string, data any) {
	writeJSON(w, http.StatusOK, dto.Success(message, data))
}

// writeCreated menulis respons sukses dengan status 201.
func writeCreated(w http.ResponseWriter, message string, data any) {
	writeJSON(w, http.StatusCreated, dto.Success(message, data))
}

// writeError memetakan error service ke status HTTP yang sesuai.
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

// decodeJSON membaca dan memvalidasi body JSON request (maks 1 MB, tanpa field tak dikenal).
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

// pathID mengambil path parameter "id" sebagai uint.
func pathID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, service.ErrBadRequest
	}
	return uint(id), nil
}

// queryUint membaca query parameter uint dengan nilai default.
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

// queryInt membaca query parameter integer dengan nilai default.
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

// queryFloat membaca query parameter float dengan nilai default.
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
