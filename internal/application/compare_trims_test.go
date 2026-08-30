package application

import (
	"context"
	"testing"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type fakeTrimGetter struct {
	byID map[int]domain.TrimDetail
}

func (f fakeTrimGetter) Execute(_ context.Context, id int) (domain.TrimDetail, error) {
	detail, ok := f.byID[id]
	if !ok {
		return domain.TrimDetail{}, domain.ErrNotFound
	}
	return detail, nil
}

type fakeMarketGetter struct {
	median int64
	count  int
}

func (f fakeMarketGetter) Execute(context.Context, int, int) (domain.TrimMarketSummary, error) {
	median := f.median
	return domain.TrimMarketSummary{Count: f.count, Median: &median, Currency: "ARS"}, nil
}

func TestCompareTrims_featureMatrixFillsDash(t *testing.T) {
	hp := 170
	uc := NewCompareTrims(
		fakeTrimGetter{byID: map[int]domain.TrimDetail{
			1: {
				TrimID: 1, TrimName: "XEi", BrandName: "Toyota", ModelName: "Corolla",
				Specs:    &domain.VehicleSpec{Engine: "2.0", Transmission: "CVT", Horsepower: &hp},
				Features: []domain.VehicleFeature{{Code: "airbags", Name: "Airbags", Category: "safety", Value: "7"}},
			},
			2: {
				TrimID: 2, TrimName: "XLi", BrandName: "Toyota", ModelName: "Corolla",
				Specs:    &domain.VehicleSpec{Engine: "1.8", Transmission: "CVT"},
				Features: []domain.VehicleFeature{{Code: "esp", Name: "ESP", Category: "safety", Value: "true"}},
			},
		}},
		fakeMarketGetter{median: 25_000_000, count: 10},
	)

	got, err := uc.Execute(context.Background(), []int{1, 2}, 2019)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Trims) != 2 {
		t.Fatalf("trims = %d", len(got.Trims))
	}
	if got.Trims[0].Specs == nil || got.Trims[0].Specs.Engine != "2.0" {
		t.Fatalf("trim 1 specs = %+v", got.Trims[0].Specs)
	}
	if got.Trims[1].Specs == nil || got.Trims[1].Specs.Engine != "1.8" {
		t.Fatalf("trim 2 specs = %+v", got.Trims[1].Specs)
	}

	byCode := map[string]CompareFeature{}
	for _, f := range got.Features {
		byCode[f.Code] = f
	}
	if byCode["airbags"].Values["1"] != "7" || byCode["airbags"].Values["2"] != "—" {
		t.Fatalf("airbags values = %v", byCode["airbags"].Values)
	}
	if byCode["esp"].Values["2"] != "true" || byCode["esp"].Values["1"] != "—" {
		t.Fatalf("esp values = %v", byCode["esp"].Values)
	}
}
