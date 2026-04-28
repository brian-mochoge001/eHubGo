package handlers

import (
	"math"
)

// RoundToNearestTen rounds the given amount to the nearest 10.
func RoundToNearestTen(amount float64) float64 {
	return math.Round(amount/10.0) * 10.0
}

// CalculatePrice computes the final price based on base, distance, and surcharges.
func CalculatePrice(baseFee, ratePerKm, distance, peakMultiplier, weatherSurcharge float64) float64 {
	baseCost := baseFee + (ratePerKm * distance)
	total := (baseCost * peakMultiplier) + weatherSurcharge
	return RoundToNearestTen(total)
}
