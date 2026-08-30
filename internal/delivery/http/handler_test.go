package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arangue/autocompare-ar/internal/application"
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
	byID   map[int]domain.TrimDetail
}

func (s stubGetTrim) Execute(_ context.Context, id int) (domain.TrimDetail, error) {
	if s.err != nil {
		return domain.TrimDetail{}, s.err
	}
	if s.byID != nil {
		detail, ok := s.byID[id]
		if !ok {
			return domain.TrimDetail{}, domain.ErrNotFound
		}
		return detail, nil
	}
	return s.detail, nil
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

type stubAssessDeal struct{}

func (stubAssessDeal) Execute(context.Context, int, int, int64) (domain.DealAssessment, error) {
	return domain.DealAssessment{Currency: "ARS", Status: domain.DealStatusInsufficientData, Band: domain.DealBandInsufficientData}, nil
}

type stubCompareTrims struct {
	result application.CompareResult
	err    error
}

func (s stubCompareTrims) Execute(context.Context, []int, int) (application.CompareResult, error) {
	return s.result, s.err
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
		stubAssessDeal{},
		stubCompareTrims{},
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
		stubAssessDeal{},
		stubCompareTrims{},
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
		stubAssessDeal{},
		stubCompareTrims{},
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

func TestAssessDeal_missingPrice(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trims/1/deal?year=2019", nil)
	req.SetPathValue("trim_id", "1")
	rec := httptest.NewRecorder()
	h.AssessDeal(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var errBody errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Code != "INVALID_PRICE" {
		t.Fatalf("code = %q", errBody.Code)
	}
}

func TestCompareTrims_invalidIDs(t *testing.T) {
	h := newTestHandler()
	for _, raw := range []string{"1", "1,2,3,4", "1,x", ""} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/compare?trim_ids="+raw+"&year=2019", nil)
		rec := httptest.NewRecorder()
		h.CompareTrims(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("trim_ids=%q status = %d", raw, rec.Code)
		}
		var errBody errorBody
		if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
			t.Fatal(err)
		}
		if errBody.Code != "INVALID_TRIM_IDS" {
			t.Fatalf("trim_ids=%q code = %q", raw, errBody.Code)
		}
	}
}

func TestCompareTrims_notFound(t *testing.T) {
	compare := application.NewCompareTrims(stubGetTrim{err: domain.ErrNotFound}, stubGetMarketSummary{})
	h := NewHandler(
		stubListBrands{},
		stubListModels{},
		stubSearchTrims{},
		stubGetTrim{},
		stubListListings{},
		stubGetMarketSummary{},
		stubListReferences{},
		stubAssessDeal{},
		compare,
		stubPinger{},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/compare?trim_ids=1,99&year=2019", nil)
	rec := httptest.NewRecorder()
	h.CompareTrims(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCompareTrims_ok(t *testing.T) {
	hp := 170
	getTrim := stubGetTrim{byID: map[int]domain.TrimDetail{
		1: {
			TrimID: 1, TrimName: "XEi 2.0 CVT", BrandName: "Toyota", ModelName: "Corolla",
			Specs:    &domain.VehicleSpec{Engine: "2.0", Transmission: "CVT", Horsepower: &hp},
			Features: []domain.VehicleFeature{{Code: "airbags", Name: "Airbags", Category: "safety", Value: "7"}},
		},
		2: {
			TrimID: 2, TrimName: "XLi 1.8 CVT", BrandName: "Toyota", ModelName: "Corolla",
			Specs:    &domain.VehicleSpec{Engine: "1.8", Transmission: "CVT"},
			Features: []domain.VehicleFeature{{Code: "esp", Name: "ESP", Category: "safety", Value: "true"}},
		},
	}}
	compare := application.NewCompareTrims(getTrim, stubGetMarketSummary{})
	h := NewHandler(
		stubListBrands{},
		stubListModels{},
		stubSearchTrims{},
		getTrim,
		stubListListings{},
		stubGetMarketSummary{},
		stubListReferences{},
		stubAssessDeal{},
		compare,
		stubPinger{},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/compare?trim_ids=1,2&year=2019", nil)
	rec := httptest.NewRecorder()
	h.CompareTrims(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	trims, ok := body["trims"].([]any)
	if !ok || len(trims) != 2 {
		t.Fatalf("trims = %v", body["trims"])
	}
	features, ok := body["features"].([]any)
	if !ok || len(features) != 2 {
		t.Fatalf("features = %v", body["features"])
	}
}

func TestAssessDeal_missingYear(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trims/1/deal?price=21500000", nil)
	req.SetPathValue("trim_id", "1")
	rec := httptest.NewRecorder()
	h.AssessDeal(rec, req)

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
