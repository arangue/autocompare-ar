package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/arangue/autocompare-ar/internal/application"
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

type listListingsUseCase interface {
	Execute(ctx context.Context, trimID, year, limit int) ([]domain.Listing, error)
}

type getMarketSummaryUseCase interface {
	Execute(ctx context.Context, trimID, year int) (domain.TrimMarketSummary, error)
}

type listReferencesUseCase interface {
	Execute(ctx context.Context, trimID, year int) ([]domain.ReferencePrice, error)
}

type assessDealUseCase interface {
	Execute(ctx context.Context, trimID, year int, price int64) (domain.DealAssessment, error)
}

type compareTrimsUseCase interface {
	Execute(ctx context.Context, trimIDs []int, year int) (application.CompareResult, error)
}

type Handler struct {
	listBrands        listBrandsUseCase
	listModelsByBrand listModelsByBrandUseCase
	searchTrims       searchTrimsUseCase
	getTrim           getTrimUseCase
	listListings      listListingsUseCase
	getMarketSummary  getMarketSummaryUseCase
	listReferences    listReferencesUseCase
	assessDeal        assessDealUseCase
	compareTrims      compareTrimsUseCase
	db                pinger
}

func NewHandler(listBrands listBrandsUseCase, listModelsByBrand listModelsByBrandUseCase, searchTrims searchTrimsUseCase, getTrim getTrimUseCase, listListings listListingsUseCase, getMarketSummary getMarketSummaryUseCase, listReferences listReferencesUseCase, assessDeal assessDealUseCase, compareTrims compareTrimsUseCase, db pinger) *Handler {
	return &Handler{listBrands: listBrands, listModelsByBrand: listModelsByBrand, searchTrims: searchTrims, getTrim: getTrim, listListings: listListings, getMarketSummary: getMarketSummary, listReferences: listReferences, assessDeal: assessDeal, compareTrims: compareTrims, db: db}
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

func (h *Handler) ListListings(w http.ResponseWriter, r *http.Request) {
	trimID, err := strconv.Atoi(r.PathValue("trim_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "trim_id must be an integer")
		return
	}

	rawYear := r.URL.Query().Get("year")
	if rawYear == "" {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year is required")
		return
	}
	year, err := strconv.Atoi(rawYear)
	if err != nil || year <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year must be a positive integer")
		return
	}

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			writeError(w, http.StatusBadRequest, "INVALID_LIMIT", "limit must be a positive integer")
			return
		}
		limit = n
	}
	if limit > 100 {
		limit = 100
	}

	listings, err := h.listListings.Execute(r.Context(), trimID, year, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list listings")
		return
	}
	writeJSON(w, http.StatusOK, listings)
}

func (h *Handler) GetMarketSummary(w http.ResponseWriter, r *http.Request) {
	trimID, err := strconv.Atoi(r.PathValue("trim_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "trim_id must be an integer")
		return
	}

	rawYear := r.URL.Query().Get("year")
	if rawYear == "" {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year is required")
		return
	}
	year, err := strconv.Atoi(rawYear)
	if err != nil || year <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year must be a positive integer")
		return
	}

	summary, err := h.getMarketSummary.Execute(r.Context(), trimID, year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get market summary")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) ListReferences(w http.ResponseWriter, r *http.Request) {
	trimID, err := strconv.Atoi(r.PathValue("trim_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "trim_id must be an integer")
		return
	}

	rawYear := r.URL.Query().Get("year")
	if rawYear == "" {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year is required")
		return
	}
	year, err := strconv.Atoi(rawYear)
	if err != nil || year <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year must be a positive integer")
		return
	}

	references, err := h.listReferences.Execute(r.Context(), trimID, year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list references")
		return
	}
	writeJSON(w, http.StatusOK, references)
}

func (h *Handler) AssessDeal(w http.ResponseWriter, r *http.Request) {
	trimID, err := strconv.Atoi(r.PathValue("trim_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "trim_id must be an integer")
		return
	}

	rawYear := r.URL.Query().Get("year")
	if rawYear == "" {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year is required")
		return
	}
	year, err := strconv.Atoi(rawYear)
	if err != nil || year <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year must be a positive integer")
		return
	}

	rawPrice := r.URL.Query().Get("price")
	if rawPrice == "" {
		writeError(w, http.StatusBadRequest, "INVALID_PRICE", "price is required")
		return
	}
	price, err := strconv.ParseInt(rawPrice, 10, 64)
	if err != nil || price <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_PRICE", "price must be a positive integer")
		return
	}

	assessment, err := h.assessDeal.Execute(r.Context(), trimID, year, price)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to assess deal")
		return
	}
	writeJSON(w, http.StatusOK, assessment)
}

func (h *Handler) CompareTrims(w http.ResponseWriter, r *http.Request) {
	rawYear := r.URL.Query().Get("year")
	if rawYear == "" {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year is required")
		return
	}
	year, err := strconv.Atoi(rawYear)
	if err != nil || year <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year must be a positive integer")
		return
	}

	ids, ok := parseTrimIDs(r.URL.Query().Get("trim_ids"))
	if !ok {
		writeError(w, http.StatusBadRequest, "INVALID_TRIM_IDS", "trim_ids must be 2 or 3 comma-separated integers")
		return
	}

	result, err := h.compareTrims.Execute(r.Context(), ids, year)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "trim not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to compare trims")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func parseTrimIDs(raw string) ([]int, bool) {
	parts := strings.Split(raw, ",")
	if len(parts) < 2 || len(parts) > 3 {
		return nil, false
	}
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n <= 0 {
			return nil, false
		}
		ids = append(ids, n)
	}
	return ids, true
}
