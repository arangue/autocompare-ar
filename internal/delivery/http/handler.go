package http

import (
	"context"
	"net/http"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type pinger interface {
	Ping(ctx context.Context) error
}

type listBrandsUseCase interface {
	Execute(ctx context.Context) ([]domain.Brand, error)
}

type Handler struct {
	listBrands listBrandsUseCase
	db         pinger
}

func NewHandler(listBrands listBrandsUseCase, db pinger) *Handler {
	return &Handler{listBrands: listBrands, db: db}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListBrands(w http.ResponseWriter, r *http.Request) {
	brands, err := h.listBrands.Execute(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list brands")
		return
	}
	writeJSON(w, http.StatusOK, brands)
}
