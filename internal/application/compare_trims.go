package application

import (
	"context"
	"strconv"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type trimGetter interface {
	Execute(ctx context.Context, id int) (domain.TrimDetail, error)
}

type marketGetter interface {
	Execute(ctx context.Context, trimID, year int) (domain.TrimMarketSummary, error)
}

type CompareSpecs struct {
	Engine       string `json:"engine"`
	Transmission string `json:"transmission"`
	Horsepower   *int   `json:"horsepower"`
}

type CompareMarket struct {
	Count    int    `json:"count"`
	Median   *int64 `json:"median"`
	Currency string `json:"currency"`
}

type CompareTrim struct {
	TrimID    int           `json:"trim_id"`
	TrimName  string        `json:"trim_name"`
	BrandName string        `json:"brand_name"`
	ModelName string        `json:"model_name"`
	Specs     *CompareSpecs `json:"specs"`
	Market    CompareMarket `json:"market"`
}

type CompareFeature struct {
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	Category string            `json:"category"`
	Values   map[string]string `json:"values"`
}

type CompareResult struct {
	Year     int              `json:"year"`
	Trims    []CompareTrim    `json:"trims"`
	Features []CompareFeature `json:"features"`
}

type CompareTrims struct {
	getTrim   trimGetter
	getMarket marketGetter
}

func NewCompareTrims(getTrim trimGetter, getMarket marketGetter) *CompareTrims {
	return &CompareTrims{getTrim: getTrim, getMarket: getMarket}
}

func (u *CompareTrims) Execute(ctx context.Context, trimIDs []int, year int) (CompareResult, error) {
	trims := make([]CompareTrim, 0, len(trimIDs))
	features := make([]CompareFeature, 0)
	index := make(map[string]int)

	for _, id := range trimIDs {
		detail, err := u.getTrim.Execute(ctx, id)
		if err != nil {
			return CompareResult{}, err
		}
		summary, err := u.getMarket.Execute(ctx, id, year)
		if err != nil {
			return CompareResult{}, err
		}

		var specs *CompareSpecs
		if detail.Specs != nil {
			specs = &CompareSpecs{
				Engine:       detail.Specs.Engine,
				Transmission: detail.Specs.Transmission,
				Horsepower:   detail.Specs.Horsepower,
			}
		}

		trims = append(trims, CompareTrim{
			TrimID:    detail.TrimID,
			TrimName:  detail.TrimName,
			BrandName: detail.BrandName,
			ModelName: detail.ModelName,
			Specs:     specs,
			Market: CompareMarket{
				Count:    summary.Count,
				Median:   summary.Median,
				Currency: summary.Currency,
			},
		})

		key := strconv.Itoa(id)
		for _, feature := range detail.Features {
			if i, ok := index[feature.Code]; ok {
				features[i].Values[key] = feature.Value
				continue
			}
			index[feature.Code] = len(features)
			features = append(features, CompareFeature{
				Code:     feature.Code,
				Name:     feature.Name,
				Category: feature.Category,
				Values:   map[string]string{key: feature.Value},
			})
		}
	}

	for i := range features {
		for _, id := range trimIDs {
			key := strconv.Itoa(id)
			if _, ok := features[i].Values[key]; !ok {
				features[i].Values[key] = "—"
			}
		}
	}

	return CompareResult{Year: year, Trims: trims, Features: features}, nil
}
