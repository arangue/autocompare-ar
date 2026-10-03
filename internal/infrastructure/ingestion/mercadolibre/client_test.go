package mercadolibre

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSearch(t *testing.T) {
	var gotAuth, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"results": [
				{
					"id": "MLA100",
					"title": "Toyota Corolla XEI",
					"price": 20000000,
					"currency_id": "ARS",
					"permalink": "https://auto.mercadolibre.com.ar/MLA-100",
					"attributes": [
						{"id": "VEHICLE_YEAR", "value_name": "2019"},
						{"id": "KILOMETERS", "value_name": "90000 km"}
					]
				},
				{"id": "MLA101", "title": "sin año", "price": 10}
			]
		}`))
	}))
	defer srv.Close()

	c := New("test-token", "MLA")
	c.baseURL = srv.URL
	c.minInterval = 0

	listings, skipped, err := c.Search(context.Background(), "corolla xei", 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("auth = %q", gotAuth)
	}
	for _, want := range []string{"category=MLA1744", "condition=used", "q=corolla+xei", "limit=2"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query %q missing %q", gotQuery, want)
		}
	}
	if skipped != 1 || len(listings) != 1 {
		t.Fatalf("listings=%d skipped=%d", len(listings), skipped)
	}
	if listings[0].ExternalID != "MLA100" || listings[0].TrimID != 0 || listings[0].Year != 2019 {
		t.Fatalf("listing = %+v", listings[0])
	}
}

func TestSearch_requiresToken(t *testing.T) {
	c := New("", "MLA")
	c.minInterval = 0
	_, _, err := c.Search(context.Background(), "corolla", 1, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearch_status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"forbidden","error":"forbidden","status":403}`, http.StatusForbidden)
	}))
	defer srv.Close()

	c := New("tok", "")
	c.baseURL = srv.URL
	c.minInterval = 0
	_, _, err := c.Search(context.Background(), "x", 1, 0)
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("err = %v", err)
	}
}

func TestWait(t *testing.T) {
	c := New("tok", "MLA")
	c.minInterval = 20 * time.Millisecond
	start := time.Now()
	c.mu.Lock()
	c.wait(start)
	c.wait(start)
	c.mu.Unlock()
	if time.Since(start) < 15*time.Millisecond {
		t.Fatal("second wait returned too soon")
	}
}
