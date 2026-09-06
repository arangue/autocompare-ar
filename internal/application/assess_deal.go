package application

import (
	"context"
	"math"

	"github.com/arangue/autocompare-ar/internal/domain"
)

type AssessDeal struct {
	listingRepository domain.ListingRepository
}

func NewAssessDeal(listingRepository domain.ListingRepository) *AssessDeal {
	return &AssessDeal{listingRepository: listingRepository}
}

func (u *AssessDeal) Execute(ctx context.Context, trimID, year int, price int64) (domain.DealAssessment, error) {
	summary, err := u.listingRepository.MarketSummary(ctx, trimID, year)
	if err != nil {
		return domain.DealAssessment{}, err
	}

	market := domain.DealMarketSnapshot{
		Count:  summary.Count,
		Median: summary.Median,
		P25:    summary.P25,
		P75:    summary.P75,
	}

	if summary.Count < domain.DealMinSampleSize || summary.Median == nil {
		return domain.DealAssessment{
			TrimID:     trimID,
			Year:       year,
			Price:      price,
			Currency:   "ARS",
			Status:     domain.DealStatusInsufficientData,
			Band:       domain.DealBandInsufficientData,
			Market:     market,
			Disclaimer: domain.DealDisclaimerInsufficientData,
		}, nil
	}

	deltaARS := *summary.Median - price
	deltaPct := roundDeltaPct(float64(deltaARS) / float64(*summary.Median) * 100)

	return domain.DealAssessment{
		TrimID:     trimID,
		Year:       year,
		Price:      price,
		Currency:   "ARS",
		Status:     domain.DealStatusOK,
		Band:       classifyDealBand(price, summary),
		DeltaARS:   &deltaARS,
		DeltaPct:   &deltaPct,
		Market:     market,
		Disclaimer: domain.DealDisclaimerOK,
	}, nil
}

func classifyDealBand(price int64, summary domain.MarketSummary) string {
	if summary.P25 != nil && summary.P75 != nil {
		switch {
		case price < *summary.P25:
			return domain.DealBandBelow
		case price > *summary.P75:
			return domain.DealBandAbove
		default:
			return domain.DealBandNear
		}
	}

	if summary.Median == nil {
		return domain.DealBandNear
	}

	median := float64(*summary.Median)
	switch {
	case float64(price) < median*0.95:
		return domain.DealBandBelow
	case float64(price) > median*1.05:
		return domain.DealBandAbove
	default:
		return domain.DealBandNear
	}
}

func roundDeltaPct(value float64) float64 {
	return math.Round(value*10) / 10
}
