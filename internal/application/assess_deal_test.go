package application

import (
	"context"
	"testing"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type fakeListingRepo struct {
	summary domain.MarketSummary
	err     error
}

func (f fakeListingRepo) ListByTrimYear(context.Context, int, int, int) ([]domain.Listing, error) {
	return nil, nil
}

func (f fakeListingRepo) MarketSummary(context.Context, int, int) (domain.MarketSummary, error) {
	return f.summary, f.err
}

func (f fakeListingRepo) Upsert(context.Context, domain.Listing) (bool, error) {
	return false, nil
}

func (f fakeListingRepo) ExpireStale(context.Context, int) (int64, error) {
	return 0, nil
}

func int64Ptr(v int64) *int64 { return &v }

func TestAssessDeal_insufficientData(t *testing.T) {
	uc := NewAssessDeal(fakeListingRepo{summary: domain.MarketSummary{Count: 0}})

	got, err := uc.Execute(context.Background(), 1, 2019, 21_500_000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.DealStatusInsufficientData {
		t.Fatalf("status = %q", got.Status)
	}
	if got.Band != domain.DealBandInsufficientData {
		t.Fatalf("band = %q", got.Band)
	}
	if got.DeltaARS != nil || got.DeltaPct != nil {
		t.Fatalf("expected null deltas, got delta_ars=%v delta_pct=%v", got.DeltaARS, got.DeltaPct)
	}
}

func TestAssessDeal_belowMedian(t *testing.T) {
	uc := NewAssessDeal(fakeListingRepo{summary: domain.MarketSummary{
		Count:  12,
		Median: int64Ptr(26_000_000),
		P25:    int64Ptr(25_200_000),
		P75:    int64Ptr(27_100_000),
	}})

	got, err := uc.Execute(context.Background(), 1, 2019, 21_500_000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.DealStatusOK {
		t.Fatalf("status = %q", got.Status)
	}
	if got.Band != domain.DealBandBelow {
		t.Fatalf("band = %q", got.Band)
	}
	if got.DeltaARS == nil || *got.DeltaARS != 4_500_000 {
		t.Fatalf("delta_ars = %v", got.DeltaARS)
	}
	if got.DeltaPct == nil || *got.DeltaPct != 17.3 {
		t.Fatalf("delta_pct = %v", got.DeltaPct)
	}
}

func TestAssessDeal_nearMedian(t *testing.T) {
	uc := NewAssessDeal(fakeListingRepo{summary: domain.MarketSummary{
		Count:  10,
		Median: int64Ptr(26_000_000),
		P25:    int64Ptr(25_200_000),
		P75:    int64Ptr(27_100_000),
	}})

	got, err := uc.Execute(context.Background(), 1, 2019, 26_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Band != domain.DealBandNear {
		t.Fatalf("band = %q", got.Band)
	}
}

func TestAssessDeal_aboveMedian(t *testing.T) {
	uc := NewAssessDeal(fakeListingRepo{summary: domain.MarketSummary{
		Count:  10,
		Median: int64Ptr(26_000_000),
		P25:    int64Ptr(25_200_000),
		P75:    int64Ptr(27_100_000),
	}})

	got, err := uc.Execute(context.Background(), 1, 2019, 28_500_000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Band != domain.DealBandAbove {
		t.Fatalf("band = %q", got.Band)
	}
	if got.DeltaARS == nil || *got.DeltaARS >= 0 {
		t.Fatalf("delta_ars = %v, want negative", got.DeltaARS)
	}
}

func TestAssessDeal_fallbackWithoutPercentiles(t *testing.T) {
	uc := NewAssessDeal(fakeListingRepo{summary: domain.MarketSummary{
		Count:  5,
		Median: int64Ptr(20_000_000),
	}})

	got, err := uc.Execute(context.Background(), 1, 2019, 18_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Band != domain.DealBandBelow {
		t.Fatalf("band = %q", got.Band)
	}
}

func TestAssessDeal_countBelowMinIsInsufficient(t *testing.T) {
	uc := NewAssessDeal(fakeListingRepo{summary: domain.MarketSummary{
		Count:  4,
		Median: int64Ptr(26_000_000),
	}})
	got, err := uc.Execute(context.Background(), 1, 2019, 21_500_000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.DealStatusInsufficientData {
		t.Fatalf("status = %q", got.Status)
	}
	if got.DeltaARS != nil || got.DeltaPct != nil {
		t.Fatalf("deltas must be nil")
	}
}

func TestAssessDeal_minSampleIsOK(t *testing.T) {
	uc := NewAssessDeal(fakeListingRepo{summary: domain.MarketSummary{
		Count:  5,
		Median: int64Ptr(26_000_000),
		P25:    int64Ptr(25_200_000),
		P75:    int64Ptr(27_100_000),
	}})
	got, err := uc.Execute(context.Background(), 1, 2019, 21_500_000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.DealStatusOK {
		t.Fatalf("status = %q", got.Status)
	}
}
