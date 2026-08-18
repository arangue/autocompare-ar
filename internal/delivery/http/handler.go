package http

import (
	"context"
	"net/http"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type listBrandsUseCase interface {
	Execute(ctx context.Context) ([]domain.Brand, error)
}

type Handler struct {
	listBrands listBrandsUseCase
}

func NewHandler(listBrands listBrandsUseCase) *Handler {
	return &Handler{listBrands: listBrands}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
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
