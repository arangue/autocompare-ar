package http

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /api/v1/brands", h.ListBrands)
	mux.HandleFunc("GET /api/v1/brands/{brand_id}/models", h.ListModelsByBrand)
	mux.HandleFunc("GET /api/v1/search/trims", h.SearchTrims)
	mux.HandleFunc("GET /api/v1/trims/{trim_id}", h.GetTrim)
	return mux
}
