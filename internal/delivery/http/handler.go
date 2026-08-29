package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type pinger interface {
	Ping(ctx context.Context) error
}

type listModelsByBrandUseCase interface {
	Execute(ctx context.Context, brandID int) ([]domain.Model, error)
}

type listBrandsUseCase interface {
	Execute(ctx context.Context) ([]domain.Brand, error)
}

type searchTrimsUseCase interface {
	Execute(ctx context.Context, query string, limit int) ([]domain.TrimSearchResult, error)
}

type getTrimUseCase interface {
	Execute(ctx context.Context, id int) (domain.TrimDetail, error)
}

type Handler struct {
	listBrands        listBrandsUseCase
	listModelsByBrand listModelsByBrandUseCase
	searchTrims       searchTrimsUseCase
	getTrim           getTrimUseCase
	db                pinger
}

func NewHandler(listBrands listBrandsUseCase, listModelsByBrand listModelsByBrandUseCase, searchTrims searchTrimsUseCase, getTrim getTrimUseCase, db pinger) *Handler {
	return &Handler{listBrands: listBrands, listModelsByBrand: listModelsByBrand, searchTrims: searchTrims, getTrim: getTrim, db: db}
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

func (h *Handler) ListModelsByBrand(w http.ResponseWriter, r *http.Request) {
	brandIDInt, err := strconv.Atoi(r.PathValue("brand_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "brand_id must be an integer")
		return
	}

	models, err := h.listModelsByBrand.Execute(r.Context(), brandIDInt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list models by brand")
		return
	}
	writeJSON(w, http.StatusOK, models)
}

func (h *Handler) SearchTrims(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusOK, []domain.TrimSearchResult{})
		return
	}

	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			writeError(w, http.StatusBadRequest, "INVALID_LIMIT", "limit must be a positive integer")
			return
		}
		limit = n
	}
	if limit > 50 {
		limit = 50
	}

	trims, err := h.searchTrims.Execute(r.Context(), query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search trims")
		return
	}
	writeJSON(w, http.StatusOK, trims)
}

func (h *Handler) GetTrim(w http.ResponseWriter, r *http.Request) {
	trimIDInt, err := strconv.Atoi(r.PathValue("trim_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "trim_id must be an integer")
		return
	}

	var requestedYear *int
	if raw := r.URL.Query().Get("year"); raw != "" {
		y, err := strconv.Atoi(raw)
		if err != nil || y <= 0 {
			writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year must be a positive integer")
			return
		}
		requestedYear = &y
	}

	trim, err := h.getTrim.Execute(r.Context(), trimIDInt)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "trim not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get trim")
		return
	}
	trim.RequestedYear = requestedYear
	writeJSON(w, http.StatusOK, trim)
}
