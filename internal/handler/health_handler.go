package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/dto"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, dto.Error("database tidak terjangkau"))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("ok", map[string]any{"database": "ok"}))
}
