package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type stubPinger struct{}

func (stubPinger) Ping(context.Context) error { return nil }

type stubListBrands struct {
	brands []domain.Brand
}

func (s stubListBrands) Execute(context.Context) ([]domain.Brand, error) {
	return s.brands, nil
}

type stubListModels struct{}

func (stubListModels) Execute(context.Context, int) ([]domain.Model, error) {
	return nil, nil
}

type stubSearchTrims struct{}

func (stubSearchTrims) Execute(context.Context, string, int) ([]domain.TrimSearchResult, error) {
	return nil, nil
}

type stubGetTrim struct {
	detail domain.TrimDetail
	err    error
}

func (s stubGetTrim) Execute(context.Context, int) (domain.TrimDetail, error) {
	return s.detail, s.err
}

type stubListListings struct {
	listings []domain.Listing
}

func (s stubListListings) Execute(context.Context, int, int, int) ([]domain.Listing, error) {
	return s.listings, nil
}

type stubGetMarketSummary struct{}

func (stubGetMarketSummary) Execute(context.Context, int, int) (domain.TrimMarketSummary, error) {
	return domain.TrimMarketSummary{Currency: "ARS"}, nil
}

type stubListReferences struct{}

func (stubListReferences) Execute(context.Context, int, int) ([]domain.ReferencePrice, error) {
	return nil, nil
}

func newTestHandler() *Handler {
	return NewHandler(
		stubListBrands{},
		stubListModels{},
		stubSearchTrims{},
		stubGetTrim{},
		stubListListings{},
		stubGetMarketSummary{},
		stubListReferences{},
		stubPinger{},
	)
}

func TestListBrands_snakeCaseJSON(t *testing.T) {
	h := NewHandler(
		stubListBrands{brands: []domain.Brand{{ID: 1, Name: "Toyota", CreatedAt: time.Unix(0, 0).UTC()}}},
		stubListModels{},
		stubSearchTrims{},
		stubGetTrim{},
		stubListListings{},
		stubGetMarketSummary{},
		stubListReferences{},
		stubPinger{},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/brands", nil)
	rec := httptest.NewRecorder()
	h.ListBrands(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body[0]["id"]; !ok {
		t.Fatalf("expected snake_case id, got %v", body[0])
	}
	if _, ok := body[0]["ID"]; ok {
		t.Fatalf("unexpected PascalCase ID in %v", body[0])
	}
}

func TestGetTrim_invalidYear(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trims/1?year=0", nil)
	req.SetPathValue("trim_id", "1")
	rec := httptest.NewRecorder()
	h.GetTrim(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var errBody errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Code != "INVALID_YEAR" {
		t.Fatalf("code = %q", errBody.Code)
	}
}

func TestGetTrim_notFound(t *testing.T) {
	h := NewHandler(
		stubListBrands{},
		stubListModels{},
		stubSearchTrims{},
		stubGetTrim{err: domain.ErrNotFound},
		stubListListings{},
		stubGetMarketSummary{},
		stubListReferences{},
		stubPinger{},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trims/99", nil)
	req.SetPathValue("trim_id", "99")
	rec := httptest.NewRecorder()
	h.GetTrim(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListListings_missingYear(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trims/1/listings", nil)
	req.SetPathValue("trim_id", "1")
	rec := httptest.NewRecorder()
	h.ListListings(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var errBody errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Code != "INVALID_YEAR" {
		t.Fatalf("code = %q", errBody.Code)
	}
}

func TestGetMarketSummary_missingYear(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trims/1/market", nil)
	req.SetPathValue("trim_id", "1")
	rec := httptest.NewRecorder()
	h.GetMarketSummary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var errBody errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Code != "INVALID_YEAR" {
		t.Fatalf("code = %q", errBody.Code)
	}
}
