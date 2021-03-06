package battle

import (
	"math/rand"

	"github.com/pkg/errors"
)

func random(min, max float64) float64 {
	if min > max {
		panic(errors.Errorf("min (%.1f) greater max (%.1f)", min, max))
	}
	return min + rand.Float64()*(max-min)
}

// luck returns true if we had luck given the chance percentage
func luck(chance float64) bool {
	if chance > 1 || chance < 0 {
		panic(errors.Errorf("invalid chance value: %.1f", chance))
	}
	return random(0, 1) < chance
}
